package room

import (
	"bingosync/internal/game"
	"bingosync/internal/user"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
)

var (
	ErrRoomNotFound      = errors.New("room not found")
	ErrRoomFull          = errors.New("room is full")
	ErrWrongPassword     = errors.New("wrong password")
	ErrNotOwner          = errors.New("only room owner can do this")
	ErrGameInProgress    = errors.New("game in progress")
	ErrUserNotFound      = errors.New("user not found")
	ErrPlayerAlreadySet  = errors.New("player already set for this color")
)

// Room represents a game room
type Room struct {
	mu          sync.RWMutex
	ID          string
	Name        string
	Password    string
	OwnerID     string
	Game        *game.Game
	Users       map[string]*user.User
	UserOrder   []string // Order of users for reference
}

// NewRoom creates a new room
func NewRoom(id, name, password, ownerID string) *Room {
	return &Room{
		ID:        id,
		Name:      name,
		Password:  password,
		OwnerID:   ownerID,
		Game:      game.NewGame(game.RuleNormal),
		Users:     make(map[string]*user.User),
		UserOrder: []string{},
	}
}

// AddUser adds a user to the room
func (r *Room) AddUser(u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.Users[u.ID]; exists {
		return nil // Already in room
	}
	
	u.RoomID = r.ID
	u.Role = user.RoleSpectator
	u.PlayerColor = user.ColorNone
	r.Users[u.ID] = u
	r.UserOrder = append(r.UserOrder, u.ID)
	
	// First user becomes owner and referee
	if r.OwnerID == "" {
		r.OwnerID = u.ID
		u.Role = user.RoleReferee
	}
	
	return nil
}

// RemoveUser removes a user from the room
func (r *Room) RemoveUser(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if u, exists := r.Users[userID]; exists {
		u.RoomID = ""
		u.Role = user.RoleSpectator
		u.PlayerColor = user.ColorNone
		delete(r.Users, userID)
		
		// Remove from order
		for i, id := range r.UserOrder {
			if id == userID {
				r.UserOrder = append(r.UserOrder[:i], r.UserOrder[i+1:]...)
				break
			}
		}
		
		// Transfer ownership if owner left
		if r.OwnerID == userID && len(r.UserOrder) > 0 {
			r.OwnerID = r.UserOrder[0]
			if newOwner, ok := r.Users[r.OwnerID]; ok {
				newOwner.Role = user.RoleReferee
			}
		}
	}
}

// SetUserRole sets a user's role (owner can set anyone, users can set their own)
func (r *Room) SetUserRole(callerID, targetUserID string, role user.UserRole, color user.PlayerColor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Allow owner to set anyone's role, or users to set their own role
	if r.OwnerID != callerID && callerID != targetUserID {
		return ErrNotOwner
	}

	targetUser, exists := r.Users[targetUserID]
	if !exists {
		return ErrUserNotFound
	}

	// If setting as player, check if color is already taken
	if role == user.RolePlayer && color != user.ColorNone {
		for _, u := range r.Users {
			if u.ID != targetUserID && u.PlayerColor == color {
				return ErrPlayerAlreadySet
			}
		}
	}
	
	targetUser.Role = role
	if role == user.RolePlayer {
		targetUser.PlayerColor = color
	} else {
		targetUser.PlayerColor = user.ColorNone
	}
	
	return nil
}

// SetPassword sets the room password (only owner can do this)
func (r *Room) SetPassword(callerID, password string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.OwnerID != callerID {
		return ErrNotOwner
	}
	
	r.Password = password
	return nil
}

// ValidatePassword checks if the password is correct
func (r *Room) ValidatePassword(password string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if r.Password == "" {
		return true
	}
	return r.Password == password
}

// HasPassword returns whether the room has a password
func (r *Room) HasPassword() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Password != ""
}

// SetGameRule sets the game rule (only owner can do this)
func (r *Room) SetGameRule(callerID string, rule game.GameRule, config game.PhaseConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.OwnerID != callerID {
		return ErrNotOwner
	}
	
	if r.Game.Status == game.StatusPlaying {
		return ErrGameInProgress
	}
	
	r.Game = game.NewGame(rule)
	r.Game.PhaseConfig = config
	return nil
}

// StartGame starts the game (only owner can do this)
func (r *Room) StartGame(callerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.OwnerID != callerID {
		return ErrNotOwner
	}
	
	return r.Game.Start()
}

// MarkCell marks a cell in the game
func (r *Room) MarkCell(userID string, row, col int, playerColor game.PlayerColor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, exists := r.Users[userID]
	if !exists {
		return ErrUserNotFound
	}

	// Check permissions
	switch u.Role {
	case user.RoleReferee:
		// In blackout and phase rules, referee can mark as second player (not force overwrite)
		// In normal rule, referee still uses force overwrite
		if r.Game.Rule == game.RuleBlackout || r.Game.Rule == game.RulePhase {
			return r.Game.MarkCell(row, col, playerColor)
		}
		return r.Game.MarkCellForce(row, col, playerColor)
	case user.RolePlayer:
		// Can only mark own color
		if playerColor != game.PlayerColor(u.PlayerColor) {
			return errors.New("can only mark your own color")
		}
	case user.RoleSpectator:
		return errors.New("spectators cannot mark cells")
	}

	return r.Game.MarkCell(row, col, playerColor)
}

// UnmarkCell removes a mark from a cell (only referee can do this)
func (r *Room) UnmarkCell(userID string, row, col int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, exists := r.Users[userID]
	if !exists {
		return ErrUserNotFound
	}

	if u.Role != user.RoleReferee {
		return errors.New("only referee can unmark cells")
	}

	return r.Game.UnmarkCell(row, col)
}

// ClearCellMark clears a specific color mark from a cell
// For blackout and phase rules where both colors can mark the same cell
func (r *Room) ClearCellMark(userID string, row, col int, playerColor game.PlayerColor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, exists := r.Users[userID]
	if !exists {
		return ErrUserNotFound
	}

	// Only referee or the player who marked can clear
	if u.Role == user.RoleSpectator {
		return errors.New("spectators cannot clear marks")
	}

	// Players can only clear their own color
	if u.Role == user.RolePlayer {
		if playerColor != game.PlayerColor(u.PlayerColor) {
			return errors.New("can only clear your own color")
		}
	}

	return r.Game.ClearCellMark(row, col, playerColor)
}

// ResetGame resets the game board (only owner can do this)
func (r *Room) ResetGame(callerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.OwnerID != callerID {
		return ErrNotOwner
	}

	r.Game.Reset()
	return nil
}

// SetCellText sets the text of a cell (only owner can do this, only in waiting state)
func (r *Room) SetCellText(callerID string, row, col int, text string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.OwnerID != callerID {
		return ErrNotOwner
	}

	if r.Game.Status != game.StatusWaiting {
		return errors.New("can only set cell text in waiting state")
	}

	return r.Game.SetCellText(row, col, text)
}

// SetAllCellTexts sets all cell texts (only owner can do this, only in waiting state)
func (r *Room) SetAllCellTexts(callerID string, texts []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.OwnerID != callerID {
		return ErrNotOwner
	}

	if r.Game.Status != game.StatusWaiting {
		return errors.New("can only set cell text in waiting state")
	}

	return r.Game.SetAllCellTexts(texts)
}

// Settle triggers settlement for a player in phase rule
// Player can settle for themselves, or referee can settle for players
func (r *Room) Settle(callerID string, playerColor game.PlayerColor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, exists := r.Users[callerID]
	if !exists {
		return ErrUserNotFound
	}

	// Check permissions: player can settle themselves, referee can settle anyone
	switch u.Role {
	case user.RoleReferee:
		// Can settle for any player
	case user.RolePlayer:
		// Can only settle for themselves
		if game.PlayerColor(u.PlayerColor) != playerColor {
			return errors.New("can only settle for yourself")
		}
	case user.RoleSpectator:
		return errors.New("spectators cannot settle")
	}

	return r.Game.Settle(playerColor)
}

// GetState returns the current room state
func (r *Room) GetState() *RoomState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	users := make([]UserInfo, 0, len(r.Users))
	for _, u := range r.Users {
		users = append(users, UserInfo{
			ID:          u.ID,
			Name:        u.Name,
			Role:        u.Role.String(),
			PlayerColor: u.PlayerColor.String(),
		})
	}
	
	return &RoomState{
		ID:         r.ID,
		Name:       r.Name,
		OwnerID:    r.OwnerID,
		HasPassword: r.Password != "",
		Game:       r.Game,
		Users:      users,
	}
}

// UserInfo represents user info for room state
type UserInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	PlayerColor string `json:"player_color"`
}

// RoomState represents the full room state
type RoomState struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	OwnerID     string       `json:"owner_id"`
	HasPassword bool         `json:"has_password"`
	Game        *game.Game   `json:"game"`
	Users       []UserInfo   `json:"users"`
}

// Manager manages all rooms
type Manager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

// NewManager creates a new room manager
func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
	}
}

// CreateRoom creates a new room
func (m *Manager) CreateRoom(name, password, ownerID string) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	id := generateRoomID()
	room := NewRoom(id, name, password, ownerID)
	m.rooms[id] = room
	return room
}

// GetRoom retrieves a room by ID
func (m *Manager) GetRoom(id string) *Room {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rooms[id]
}

// DeleteRoom deletes a room
func (m *Manager) DeleteRoom(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rooms, id)
}

// ListRooms returns a list of all rooms
func (m *Manager) ListRooms() []RoomInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	rooms := make([]RoomInfo, 0, len(m.rooms))
	for _, r := range m.rooms {
		r.mu.RLock()
		ownerName := ""
		if owner, ok := r.Users[r.OwnerID]; ok {
			ownerName = owner.Name
		}
		rooms = append(rooms, RoomInfo{
			ID:          r.ID,
			Name:        r.Name,
			HasPassword: r.Password != "",
			PlayerCount: len(r.Users),
			OwnerName:   ownerName,
		})
		r.mu.RUnlock()
	}
	return rooms
}

// RoomInfo represents basic room info for listing
type RoomInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	HasPassword bool   `json:"has_password"`
	PlayerCount int    `json:"player_count"`
	OwnerName   string `json:"owner_name"`
}

// generateRoomID generates a random 8-character room ID
func generateRoomID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
