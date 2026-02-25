package user

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

// UserRole represents the role of a user in a room
type UserRole int

const (
	RoleSpectator UserRole = iota // Spectator - can only watch
	RolePlayer                    // Player - can only mark own color
	RoleReferee                   // Referee - can do all operations
)

func (r UserRole) String() string {
	switch r {
	case RoleReferee:
		return "referee"
	case RolePlayer:
		return "player"
	case RoleSpectator:
		return "spectator"
	default:
		return "unknown"
	}
}

func UserRoleFromString(s string) UserRole {
	switch s {
	case "referee":
		return RoleReferee
	case "player":
		return RolePlayer
	case "spectator":
		return RoleSpectator
	default:
		return RoleSpectator
	}
}

// PlayerColor represents which color a player is assigned to
type PlayerColor int

const (
	ColorNone PlayerColor = iota
	ColorRed
	ColorBlue
)

func (c PlayerColor) String() string {
	switch c {
	case ColorRed:
		return "red"
	case ColorBlue:
		return "blue"
	default:
		return "none"
	}
}

func PlayerColorFromString(s string) PlayerColor {
	switch s {
	case "red":
		return ColorRed
	case "blue":
		return ColorBlue
	default:
		return ColorNone
	}
}

// User represents a connected user
type User struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Role        UserRole     `json:"role"`
	PlayerColor PlayerColor  `json:"player_color"`
	RoomID      string       `json:"room_id,omitempty"`
}

// NewUser creates a new user with a random ID
func NewUser(name string) *User {
	return &User{
		ID:          generateID(),
		Name:        name,
		Role:        RoleSpectator,
		PlayerColor: ColorNone,
	}
}

// generateID generates a random 16-character hex string
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Manager manages all connected users
type Manager struct {
	mu    sync.RWMutex
	users map[string]*User
}

// NewManager creates a new user manager
func NewManager() *Manager {
	return &Manager{
		users: make(map[string]*User),
	}
}

// AddUser adds a user to the manager
func (m *Manager) AddUser(user *User) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
}

// RemoveUser removes a user from the manager
func (m *Manager) RemoveUser(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.users, userID)
}

// GetUser retrieves a user by ID
func (m *Manager) GetUser(userID string) *User {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.users[userID]
}

// SetUserRole sets the role of a user
func (m *Manager) SetUserRole(userID string, role UserRole) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user, ok := m.users[userID]; ok {
		user.Role = role
		return true
	}
	return false
}

// SetUserPlayerColor sets the player color of a user
func (m *Manager) SetUserPlayerColor(userID string, color PlayerColor) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user, ok := m.users[userID]; ok {
		user.PlayerColor = color
		return true
	}
	return false
}

// SetUserRoom sets the room a user is in
func (m *Manager) SetUserRoom(userID, roomID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user, ok := m.users[userID]; ok {
		user.RoomID = roomID
		return true
	}
	return false
}
