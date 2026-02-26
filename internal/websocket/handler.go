package websocket

import (
	"bingosync/internal/game"
	"bingosync/internal/room"
	"bingosync/internal/user"
	"bingosync/pkg/protocol"
	"encoding/json"
	"errors"
	"log"
	"sync"

	"github.com/lxzan/gws"
)

// Handler handles WebSocket connections
type Handler struct {
	userManager *user.Manager
	roomManager *room.Manager
	connections sync.Map // userID -> *gws.Conn
}

// NewHandler creates a new WebSocket handler
func NewHandler() *Handler {
	return &Handler{
		userManager: user.NewManager(),
		roomManager: room.NewManager(),
	}
}

// OnOpen handles new connections
func (h *Handler) OnOpen(socket *gws.Conn) {
	// Create a new user for this connection
	u := user.NewUser("Player")
	h.userManager.AddUser(u)
	h.connections.Store(u.ID, socket)
	
	// Store user ID in socket session
	socket.Session().Store("userID", u.ID)
	
	// Send welcome message with user ID
	h.sendToSocket(socket, protocol.Message{
		Type:   "connected",
		UserID: u.ID,
		Payload: mustMarshal(map[string]string{
			"user_id":   u.ID,
			"user_name": u.Name,
		}),
	})
}

// OnClose handles connection close
func (h *Handler) OnClose(socket *gws.Conn, err error) {
	userID, _ := socket.Session().Load("userID")
	if userID == nil {
		return
	}

	uid := userID.(string)

	// Remove user from room if in one
	u := h.userManager.GetUser(uid)
	if u != nil && u.RoomID != "" {
		r := h.roomManager.GetRoom(u.RoomID)
		if r != nil {
			r.RemoveUser(uid)
			// Delete room if empty
			if len(r.Users) == 0 {
				h.roomManager.DeleteRoom(r.ID)
			} else {
				h.broadcastRoomState(r)
			}
		}
	}

	// Remove user and connection
	h.userManager.RemoveUser(uid)
	h.connections.Delete(uid)
}

// OnMessage handles incoming messages
func (h *Handler) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer message.Close()
	
	userID, _ := socket.Session().Load("userID")
	if userID == nil {
		return
	}
	
	var msg protocol.Message
	if err := json.Unmarshal(message.Bytes(), &msg); err != nil {
		h.sendError(socket, 400, "invalid message format")
		return
	}
	
	msg.UserID = userID.(string)
	
	switch msg.Type {
	case protocol.MsgSetName:
		h.handleSetName(socket, &msg)
	case protocol.MsgCreateRoom:
		h.handleCreateRoom(socket, &msg)
	case protocol.MsgJoinRoom:
		h.handleJoinRoom(socket, &msg)
	case protocol.MsgLeaveRoom:
		h.handleLeaveRoom(socket, &msg)
	case protocol.MsgSetRole:
		h.handleSetRole(socket, &msg)
	case protocol.MsgListRooms:
		h.handleListRooms(socket)
	case protocol.MsgSetPassword:
		h.handleSetPassword(socket, &msg)
	case protocol.MsgSetRule:
		h.handleSetRule(socket, &msg)
	case protocol.MsgStartGame:
		h.handleStartGame(socket, &msg)
	case protocol.MsgMarkCell:
		h.handleMarkCell(socket, &msg)
	case protocol.MsgUnmarkCell:
		h.handleUnmarkCell(socket, &msg)
	case protocol.MsgClearCellMark:
		h.handleClearCellMark(socket, &msg)
	case protocol.MsgResetGame:
		h.handleResetGame(socket, &msg)
	case protocol.MsgSetCellText:
		h.handleSetCellText(socket, &msg)
	case protocol.MsgSettle:
		h.handleSettle(socket, &msg)
	default:
		h.sendError(socket, 400, "unknown message type")
	}
}

// OnPing handles ping
func (h *Handler) OnPing(socket *gws.Conn, payload []byte) {
	socket.WritePong(payload)
}

// OnPong handles pong
func (h *Handler) OnPong(socket *gws.Conn, payload []byte) {}

// handleSetName handles setting user name
func (h *Handler) handleSetName(socket *gws.Conn, msg *protocol.Message) {
	var payload protocol.SetNamePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.sendError(socket, 400, "invalid payload")
		return
	}

	if payload.Name == "" {
		h.sendError(socket, 400, "name cannot be empty")
		return
	}

	u := h.userManager.GetUser(msg.UserID)
	if u == nil {
		h.sendError(socket, 404, "user not found")
		return
	}

	// Cannot change name while in room
	if u.RoomID != "" {
		h.sendError(socket, 403, "cannot change name while in a room")
		return
	}

	u.Name = payload.Name

	// Send confirmation
	h.sendToSocket(socket, protocol.Message{
		Type: "name_set",
		Payload: mustMarshal(map[string]string{
			"user_name": u.Name,
		}),
	})
}

// handleCreateRoom handles room creation
func (h *Handler) handleCreateRoom(socket *gws.Conn, msg *protocol.Message) {
	var payload protocol.CreateRoomPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.sendError(socket, 400, "invalid payload")
		return
	}
	
	u := h.userManager.GetUser(msg.UserID)
	if u == nil {
		h.sendError(socket, 404, "user not found")
		return
	}
	
	// Leave current room if in one
	if u.RoomID != "" {
		oldRoom := h.roomManager.GetRoom(u.RoomID)
		if oldRoom != nil {
			oldRoom.RemoveUser(msg.UserID)
		}
	}
	
	r := h.roomManager.CreateRoom(payload.Name, payload.Password, msg.UserID)
	r.AddUser(u)
	
	// Send state update in correct format
	state := r.GetState()
	h.sendToSocket(socket, protocol.Message{
		Type:   protocol.MsgJoined,
		RoomID: r.ID,
		Payload: mustMarshal(protocol.StateUpdatePayload{
			Room: protocol.RoomPayload{
				ID:          state.ID,
				Name:        state.Name,
				OwnerID:     state.OwnerID,
				HasPassword: state.HasPassword,
			},
			Game:        convertGame(state.Game),
			Users:       convertUsers(state.Users),
			CurrentUser: msg.UserID,
		}),
	})
}

// handleJoinRoom handles joining a room
func (h *Handler) handleJoinRoom(socket *gws.Conn, msg *protocol.Message) {
	var payload protocol.JoinRoomPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.sendError(socket, 400, "invalid payload")
		return
	}
	
	u := h.userManager.GetUser(msg.UserID)
	if u == nil {
		h.sendError(socket, 404, "user not found")
		return
	}
	
	r := h.roomManager.GetRoom(payload.RoomID)
	if r == nil {
		h.sendError(socket, 404, "room not found")
		return
	}
	
	if !r.ValidatePassword(payload.Password) {
		h.sendError(socket, 403, "wrong password")
		return
	}
	
	// Leave current room if in one
	if u.RoomID != "" && u.RoomID != payload.RoomID {
		oldRoom := h.roomManager.GetRoom(u.RoomID)
		if oldRoom != nil {
			oldRoom.RemoveUser(msg.UserID)
		}
	}
	
	r.AddUser(u)
	
	// Broadcast to room
	h.broadcastRoomState(r)
}

// handleLeaveRoom handles leaving a room
func (h *Handler) handleLeaveRoom(socket *gws.Conn, msg *protocol.Message) {
	u := h.userManager.GetUser(msg.UserID)
	if u == nil || u.RoomID == "" {
		return
	}
	
	r := h.roomManager.GetRoom(u.RoomID)
	if r == nil {
		return
	}
	
	r.RemoveUser(msg.UserID)
	
	// Delete room if empty
	if len(r.Users) == 0 {
		h.roomManager.DeleteRoom(r.ID)
	} else {
		h.broadcastRoomState(r)
	}
	
	h.sendToSocket(socket, protocol.Message{
		Type: protocol.MsgLeft,
		Payload: mustMarshal(map[string]string{
			"room_id": r.ID,
		}),
	})
}

// handleSetRole handles setting user role
func (h *Handler) handleSetRole(socket *gws.Conn, msg *protocol.Message) {
	var payload protocol.SetRolePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.sendError(socket, 400, "invalid payload")
		return
	}
	
	_, r, err := h.getUserAndRoom(msg.UserID)
	if err != nil {
		h.sendError(socket, 404, err.Error())
		return
	}
	
	role := user.UserRoleFromString(payload.Role)
	color := user.PlayerColorFromString(payload.PlayerColor)
	
	if err := r.SetUserRole(msg.UserID, payload.TargetUserID, role, color); err != nil {
		h.sendError(socket, 403, err.Error())
		return
	}
	
	h.broadcastRoomState(r)
}

// handleListRooms handles listing rooms
func (h *Handler) handleListRooms(socket *gws.Conn) {
	rooms := h.roomManager.ListRooms()
	h.sendToSocket(socket, protocol.Message{
		Type: protocol.MsgRoomList,
		Payload: mustMarshal(protocol.RoomListPayload{
			Rooms: convertRooms(rooms),
		}),
	})
}

// handleSetPassword handles setting room password
func (h *Handler) handleSetPassword(socket *gws.Conn, msg *protocol.Message) {
	var payload struct {
		Password string `json:"password"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.sendError(socket, 400, "invalid payload")
		return
	}
	
	_, r, err := h.getUserAndRoom(msg.UserID)
	if err != nil {
		h.sendError(socket, 404, err.Error())
		return
	}
	
	if err := r.SetPassword(msg.UserID, payload.Password); err != nil {
		h.sendError(socket, 403, err.Error())
		return
	}
	
	h.broadcastRoomState(r)
}

// handleSetRule handles setting game rule
func (h *Handler) handleSetRule(socket *gws.Conn, msg *protocol.Message) {
	var payload protocol.SetRulePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.sendError(socket, 400, "invalid payload")
		return
	}
	
	_, r, err := h.getUserAndRoom(msg.UserID)
	if err != nil {
		h.sendError(socket, 404, err.Error())
		return
	}
	
	rule := game.GameRuleFromString(payload.Rule)
	config := game.DefaultPhaseConfig()

	if len(payload.PhaseConfig.RowScores) == 5 {
		for i, v := range payload.PhaseConfig.RowScores {
			config.RowScores[i] = v
		}
	}
	if len(payload.PhaseConfig.SecondHalfScores) == 5 {
		for i, v := range payload.PhaseConfig.SecondHalfScores {
			config.SecondHalfScores[i] = v
		}
	}
	if payload.PhaseConfig.CellsPerRow > 0 {
		config.CellsPerRow = payload.PhaseConfig.CellsPerRow
	}
	if payload.PhaseConfig.UnlockThreshold > 0 {
		config.UnlockThreshold = payload.PhaseConfig.UnlockThreshold
	}
	if payload.PhaseConfig.BingoBonus > 0 {
		config.BingoBonus = payload.PhaseConfig.BingoBonus
	}
	if payload.PhaseConfig.FinalBonus > 0 {
		config.FinalBonus = payload.PhaseConfig.FinalBonus
	}
	
	if err := r.SetGameRule(msg.UserID, rule, config); err != nil {
		h.sendError(socket, 403, err.Error())
		return
	}
	
	h.broadcastRoomState(r)
}

// handleStartGame handles starting a game
func (h *Handler) handleStartGame(socket *gws.Conn, msg *protocol.Message) {
	_, r, err := h.getUserAndRoom(msg.UserID)
	if err != nil {
		h.sendError(socket, 404, err.Error())
		return
	}
	
	if err := r.StartGame(msg.UserID); err != nil {
		h.sendError(socket, 403, err.Error())
		return
	}
	
	h.broadcastRoomState(r)
}

// handleMarkCell handles marking a cell
func (h *Handler) handleMarkCell(socket *gws.Conn, msg *protocol.Message) {
	var payload protocol.MarkCellPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.sendError(socket, 400, "invalid payload")
		return
	}
	
	_, r, err := h.getUserAndRoom(msg.UserID)
	if err != nil {
		h.sendError(socket, 404, err.Error())
		return
	}
	
	color := game.PlayerColorFromString(payload.Color)
	if err := r.MarkCell(msg.UserID, payload.Row, payload.Col, color); err != nil {
		h.sendError(socket, 403, err.Error())
		return
	}

	h.broadcastRoomState(r)
}

// handleUnmarkCell handles unmarking a cell
func (h *Handler) handleUnmarkCell(socket *gws.Conn, msg *protocol.Message) {
	var payload protocol.MarkCellPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.sendError(socket, 400, "invalid payload")
		return
	}

	_, r, err := h.getUserAndRoom(msg.UserID)
	if err != nil {
		h.sendError(socket, 404, err.Error())
		return
	}

	if err := r.UnmarkCell(msg.UserID, payload.Row, payload.Col); err != nil {
		h.sendError(socket, 403, err.Error())
		return
	}

	h.broadcastRoomState(r)
}

// handleClearCellMark handles clearing a specific color mark from a cell
func (h *Handler) handleClearCellMark(socket *gws.Conn, msg *protocol.Message) {
	var payload protocol.ClearCellMarkPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.sendError(socket, 400, "invalid payload")
		return
	}

	_, r, err := h.getUserAndRoom(msg.UserID)
	if err != nil {
		h.sendError(socket, 404, err.Error())
		return
	}

	color := game.PlayerColorFromString(payload.Color)
	if err := r.ClearCellMark(msg.UserID, payload.Row, payload.Col, color); err != nil {
		h.sendError(socket, 403, err.Error())
		return
	}

	h.broadcastRoomState(r)
}

// handleResetGame handles resetting a game
func (h *Handler) handleResetGame(socket *gws.Conn, msg *protocol.Message) {
	_, r, err := h.getUserAndRoom(msg.UserID)
	if err != nil {
		h.sendError(socket, 404, err.Error())
		return
	}
	
	if err := r.ResetGame(msg.UserID); err != nil {
		h.sendError(socket, 403, err.Error())
		return
	}
	
	h.broadcastRoomState(r)
}

// handleSetCellText handles setting cell text
func (h *Handler) handleSetCellText(socket *gws.Conn, msg *protocol.Message) {
	var payload protocol.SetCellTextPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.sendError(socket, 400, "invalid payload")
		return
	}

	_, r, err := h.getUserAndRoom(msg.UserID)
	if err != nil {
		h.sendError(socket, 404, err.Error())
		return
	}

	if len(payload.Texts) > 0 {
		// Batch set
		err = r.SetAllCellTexts(msg.UserID, payload.Texts)
	} else {
		// Single set
		err = r.SetCellText(msg.UserID, payload.Row, payload.Col, payload.Text)
	}

	if err != nil {
		h.sendError(socket, 403, err.Error())
		return
	}

	h.broadcastRoomState(r)
}

// handleSettle handles settlement for phase rule
func (h *Handler) handleSettle(socket *gws.Conn, msg *protocol.Message) {
	var payload protocol.SettlePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.sendError(socket, 400, "invalid payload")
		return
	}

	_, r, err := h.getUserAndRoom(msg.UserID)
	if err != nil {
		h.sendError(socket, 404, err.Error())
		return
	}

	player := game.PlayerColorFromString(payload.Player)
	if err := r.Settle(msg.UserID, player); err != nil {
		h.sendError(socket, 403, err.Error())
		return
	}

	h.broadcastRoomState(r)
}

// sendToSocket sends a message to a socket
func (h *Handler) sendToSocket(socket *gws.Conn, msg protocol.Message) {
	data, _ := json.Marshal(msg)
	socket.WriteMessage(gws.OpcodeText, data)
}

// sendError sends an error message
func (h *Handler) sendError(socket *gws.Conn, code int, message string) {
	h.sendToSocket(socket, protocol.Message{
		Type: protocol.MsgError,
		Payload: mustMarshal(protocol.ErrorPayload{
			Code:    code,
			Message: message,
		}),
	})
}

// broadcastRoomState broadcasts the room state to all users in the room
func (h *Handler) broadcastRoomState(r *room.Room) {
	state := r.GetState()
	msg := protocol.Message{
		Type:   protocol.MsgStateUpdate,
		RoomID: r.ID,
		Payload: mustMarshal(protocol.StateUpdatePayload{
			Room: protocol.RoomPayload{
				ID:          state.ID,
				Name:        state.Name,
				OwnerID:     state.OwnerID,
				HasPassword: state.HasPassword,
			},
			Game:        convertGame(state.Game),
			Users:       convertUsers(state.Users),
			CurrentUser: "", // Will be set per user
		}),
	}
	
	for _, u := range r.Users {
		if conn, ok := h.connections.Load(u.ID); ok {
			msgCopy := msg
			payload := protocol.StateUpdatePayload{}
			json.Unmarshal(msgCopy.Payload, &payload)
			payload.CurrentUser = u.ID
			msgCopy.Payload = mustMarshal(payload)
			h.sendToSocket(conn.(*gws.Conn), msgCopy)
		}
	}
}

// Helper functions

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

// getUserAndRoom validates and returns user and room for handlers that require both
func (h *Handler) getUserAndRoom(userID string) (*user.User, *room.Room, error) {
	u := h.userManager.GetUser(userID)
	if u == nil {
		return nil, nil, errors.New("user not found")
	}
	if u.RoomID == "" {
		return nil, nil, errors.New("user not in room")
	}
	r := h.roomManager.GetRoom(u.RoomID)
	if r == nil {
		return nil, nil, errors.New("room not found")
	}
	return u, r, nil
}

func convertRooms(rooms []room.RoomInfo) []protocol.RoomPayload {
	result := make([]protocol.RoomPayload, len(rooms))
	for i, r := range rooms {
		result[i] = protocol.RoomPayload{
			ID:          r.ID,
			Name:        r.Name,
			HasPassword: r.HasPassword,
		}
	}
	return result
}

func convertGame(g *game.Game) protocol.GamePayload {
	cells := make([][]protocol.CellPayload, 5)
	for i := 0; i < 5; i++ {
		cells[i] = make([]protocol.CellPayload, 5)
		for j := 0; j < 5; j++ {
			cells[i][j] = protocol.CellPayload{
				MarkedBy:   g.Board.Cells[i][j].MarkedBy.String(),
				SecondMark: g.Board.Cells[i][j].SecondMark.String(),
				Times:      g.Board.Cells[i][j].Times,
				Text:       g.Board.Cells[i][j].Text,
			}
		}
	}

	var winner *protocol.WinnerPayload
	if g.Winner != nil {
		winner = &protocol.WinnerPayload{
			Winner:    g.Winner.Winner.String(),
			Reason:    string(g.Winner.Reason),
			RedScore:  g.Winner.RedScore,
			BlueScore: g.Winner.BlueScore,
		}
	}

	return protocol.GamePayload{
		Board: protocol.BoardPayload{
			Cells: cells,
		},
		Rule:            g.Rule.String(),
		PhaseConfig:     convertPhaseConfig(g.PhaseConfig),
		Status:          g.Status.String(),
		Winner:          winner,
		RedRowMarks:     g.RedRowMarks[:],
		BlueRowMarks:    g.BlueRowMarks[:],
		RedUnlockedRow:  g.RedUnlockedRow,
		BlueUnlockedRow: g.BlueUnlockedRow,
		BingoAchiever:   g.BingoAchiever.String(),
		BingoLine:       g.BingoLine,
		RedSettled:      g.RedSettled,
		BlueSettled:     g.BlueSettled,
		FirstSettler:    g.FirstSettler.String(),
	}
}

func convertPhaseConfig(c game.PhaseConfig) protocol.PhaseConfigPayload {
	return protocol.PhaseConfigPayload{
		RowScores:        c.RowScores[:],
		SecondHalfScores: c.SecondHalfScores[:],
		CellsPerRow:      c.CellsPerRow,
		UnlockThreshold:  c.UnlockThreshold,
		BingoBonus:       c.BingoBonus,
		FinalBonus:       c.FinalBonus,
	}
}

func convertUsers(users []room.UserInfo) []protocol.UserPayload {
	result := make([]protocol.UserPayload, len(users))
	for i, u := range users {
		result[i] = protocol.UserPayload{
			ID:          u.ID,
			Name:        u.Name,
			Role:        u.Role,
			PlayerColor: u.PlayerColor,
		}
	}
	return result
}

// GetRoomManager returns the room manager for external access
func (h *Handler) GetRoomManager() *room.Manager {
	return h.roomManager
}

// Log logs a message
func (h *Handler) Log(format string, args ...interface{}) {
	log.Printf(format, args...)
}
