package protocol

import "encoding/json"

// ProtocolVersion is the current protocol version
// Increment this when making incompatible protocol changes
const ProtocolVersion = 1

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// User operations
	MsgSetName MessageType = "set_name"

	// Room operations
	MsgCreateRoom  MessageType = "create_room"
	MsgJoinRoom    MessageType = "join_room"
	MsgLeaveRoom   MessageType = "leave_room"
	MsgSetRole     MessageType = "set_role"
	MsgListRooms   MessageType = "list_rooms"
	MsgSetPassword MessageType = "set_password"

	// Game operations
	MsgMarkCell      MessageType = "mark_cell"
	MsgUnmarkCell    MessageType = "unmark_cell"
	MsgClearCellMark MessageType = "clear_cell_mark"
	MsgSetRule       MessageType = "set_rule"
	MsgStartGame     MessageType = "start_game"
	MsgResetGame     MessageType = "reset_game"
	MsgSetCellText   MessageType = "set_cell_text"
	MsgSettle        MessageType = "settle"

	// Responses/Broadcasts
	MsgStateUpdate MessageType = "state_update"
	MsgRoomList    MessageType = "room_list"
	MsgError       MessageType = "error"
	MsgJoined      MessageType = "joined"
	MsgLeft        MessageType = "left"
)

// Message is the base message structure
type Message struct {
	Type    MessageType     `json:"type"`
	RoomID  string          `json:"room_id,omitempty"`
	UserID  string          `json:"user_id,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// CreateRoomPayload represents the payload for creating a room
type CreateRoomPayload struct {
	Name     string `json:"name"`
	Password string `json:"password,omitempty"`
	UserName string `json:"user_name"`
}

// SetNamePayload represents the payload for setting user name
type SetNamePayload struct {
	Name string `json:"name"`
}

// JoinRoomPayload represents the payload for joining a room
type JoinRoomPayload struct {
	RoomID   string `json:"room_id"`
	Password string `json:"password,omitempty"`
	UserName string `json:"user_name"`
}

// SetRolePayload represents the payload for setting a user role
type SetRolePayload struct {
	TargetUserID string `json:"target_user_id"`
	Role         string `json:"role"`
	PlayerColor  string `json:"player_color,omitempty"`
}

// MarkCellPayload represents the payload for marking a cell
type MarkCellPayload struct {
	Row   int    `json:"row"`
	Col   int    `json:"col"`
	Color string `json:"color"`
}

// ClearCellMarkPayload represents the payload for clearing a specific color mark
type ClearCellMarkPayload struct {
	Row   int    `json:"row"`
	Col   int    `json:"col"`
	Color string `json:"color"`
}

// SetRulePayload represents the payload for setting game rule
type SetRulePayload struct {
	Rule        string             `json:"rule"`
	PhaseConfig PhaseConfigPayload `json:"phase_config,omitempty"`
}

// SetCellTextPayload represents the payload for setting cell text
type SetCellTextPayload struct {
	Row   int      `json:"row,omitempty"`
	Col   int      `json:"col,omitempty"`
	Text  string   `json:"text,omitempty"`
	Texts []string `json:"texts,omitempty"`
}

// SettlePayload represents the payload for settlement (phase rule)
type SettlePayload struct {
	Player string `json:"player"`
}

// PhaseConfigPayload represents phase rule configuration
type PhaseConfigPayload struct {
	RowScores        []int `json:"row_scores"`
	SecondHalfScores []int `json:"second_half_scores"`
	CellsPerRow      int   `json:"cells_per_row"`
	UnlockThreshold  int   `json:"unlock_threshold"`
	BingoBonus       int   `json:"bingo_bonus"`
	FinalBonus       int   `json:"final_bonus"`
}

// StateUpdatePayload represents the full game state
type StateUpdatePayload struct {
	Room        RoomPayload   `json:"room"`
	Game        GamePayload   `json:"game"`
	Users       []UserPayload `json:"users"`
	CurrentUser string        `json:"current_user"`
}

// RoomPayload represents room information
type RoomPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	OwnerID     string `json:"owner_id"`
	HasPassword bool   `json:"has_password"`
}

// GamePayload represents game state
type GamePayload struct {
	Board           BoardPayload       `json:"board"`
	Rule            string             `json:"rule"`
	PhaseConfig     PhaseConfigPayload `json:"phase_config,omitempty"`
	Status          string             `json:"status"`
	Winner          *WinnerPayload     `json:"winner,omitempty"`
	RedRowMarks     []int              `json:"red_row_marks,omitempty"`
	BlueRowMarks    []int              `json:"blue_row_marks,omitempty"`
	RedUnlockedRow  int                `json:"red_unlocked_row,omitempty"`
	BlueUnlockedRow int                `json:"blue_unlocked_row,omitempty"`
	BingoAchiever   string             `json:"bingo_achiever,omitempty"`
	BingoLine       int                `json:"bingo_line,omitempty"`
	RedSettled      bool               `json:"red_settled,omitempty"`
	BlueSettled     bool               `json:"blue_settled,omitempty"`
	FirstSettler    string             `json:"first_settler,omitempty"`
}

// BoardPayload represents the board state
type BoardPayload struct {
	Cells [][]CellPayload `json:"cells"`
}

// CellPayload represents a cell state
type CellPayload struct {
	MarkedBy   string `json:"marked_by"`
	SecondMark string `json:"second_mark,omitempty"`
	Times      int    `json:"times"`
	Text       string `json:"text"`
}

// UserPayload represents user information
type UserPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	PlayerColor string `json:"player_color"`
}

// WinnerPayload represents winner information
type WinnerPayload struct {
	Winner    string `json:"winner"`
	Reason    string `json:"reason"`
	RedScore  int    `json:"red_score"`
	BlueScore int    `json:"blue_score"`
}

// RoomListPayload represents a list of rooms
type RoomListPayload struct {
	Rooms []RoomPayload `json:"rooms"`
}

// ErrorPayload represents an error message
type ErrorPayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// JoinedPayload represents a user joined notification
type JoinedPayload struct {
	User UserPayload `json:"user"`
}

// LeftPayload represents a user left notification
type LeftPayload struct {
	UserID string `json:"user_id"`
}
