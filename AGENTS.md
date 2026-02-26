# BingoSync - AI Agent Development Guide

## Project Overview

BingoSync is a real-time multiplayer Bingo game built with Go (backend) and Vue.js + TypeScript (frontend). It supports both standalone WebSocket server mode and Wails desktop application mode.

### Key Features
- Real-time multiplayer Bingo game
- Multiple game rules (Normal, Blackout, Phase)
- Role-based system (Player, Referee, Spectator)
- Room management with password protection
- Multi-language support (zh-CN, en-US)
- Light/Dark theme support
- CSV import/export for board text
- Streamer mode for OBS/broadcast

## Project Standards

### ⚠️ IMPORTANT: Language Requirements

**All code comments and documentation MUST be written in English.**

This is a strict project requirement to ensure:
- Code maintainability across international teams
- Consistency throughout the codebase
- Better accessibility for AI agents and developers

Examples:
```go
// ✅ CORRECT
// Calculate the optimal font size for a cell

// ❌ WRONG
// 计算格子的最佳字体大小
```

```typescript
// ✅ CORRECT
// Watch for dialog open, auto focus

// ❌ WRONG  
// 监听对话框打开，自动聚焦
```

### Code Style Guidelines

1. **Comments**: All comments must be in English
2. **Documentation**: All documentation files must be in English
3. **Variable Names**: Use descriptive English names (camelCase for JS/TS, PascalCase for Go exports)
4. **Commit Messages**: Prefer English for commit messages

## Tech Stack

### Backend
- **Go 1.24**
- **Wails v2** - Desktop application framework
- **gws** - WebSocket library for real-time communication

### Frontend
- **Vue 3** + **TypeScript**
- **Pinia** - State management
- **Vite** - Build tool
- **CSS Variables** - Theming system

## Project Structure

```
bingosync/
├── main.go                 # Wails application entry
├── app.go                  # Wails App bindings
├── cmd/
│   └── server/
│       └── main.go         # Standalone server entry
├── internal/
│   ├── game/
│   │   ├── game.go         # Game logic (marking, winning detection)
│   │   └── types.go        # Game types and constants
│   ├── room/
│   │   └── room.go         # Room management, user roles
│   ├── user/
│   │   └── user.go         # User entity and types
│   └── websocket/
│       └── handler.go      # WebSocket message handling
├── pkg/
│   └── protocol/
│       └── protocol.go     # Message protocol definitions
└── frontend/
    ├── src/
    │   ├── components/
    │   │   ├── BingoBoard.vue    # Main game board
    │   │   ├── PlayerPanel.vue   # Player list and roles
    │   │   ├── RoomList.vue      # Room listing
    │   │   └── RoomSettings.vue  # Room configuration
    │   ├── stores/
    │   │   ├── game.ts      # Game state (Pinia)
    │   │   ├── locale.ts    # i18n state
    │   │   └── theme.ts     # Theme state
    │   ├── locales/
    │   │   ├── zh-CN.ts     # Chinese translations
    │   │   ├── en-US.ts     # English translations
    │   │   └── index.ts     # Locale exports
    │   ├── composables/
    │   │   └── useWebSocket.ts  # WebSocket composable
    │   ├── types/
    │   │   └── index.ts     # TypeScript type definitions
    │   ├── App.vue          # Root component
    │   ├── main.ts          # Application entry
    │   └── style.css        # Global styles
    └── wailsjs/             # Wails generated bindings
```

## Architecture

### Backend Architecture

1. **Room Manager** - Manages all game rooms
2. **Room** - Contains game state, users, and owner
3. **Game** - Board state, rules, win detection
4. **User** - User info, role, player color
5. **WebSocket Handler** - Message routing and broadcasting

### Message Flow
```
Client → WebSocket → Handler → Room/Game → Broadcast → Clients
```

### Frontend Architecture

1. **Pinia Store** (`game.ts`) - Central state management
2. **WebSocket Composable** - Connection and message handling
3. **Components** - UI components consuming store state
4. **Locale Store** - i18n with reactive translations

## Code Conventions

### Go Backend

```go
// Error definitions at package level
var (
    ErrRoomNotFound = errors.New("room not found")
)

// Method receivers use short names
func (r *Room) AddUser(u *User) error {}

// Mutex naming: mu for RWMutex
type Room struct {
    mu sync.RWMutex
}

// JSON tags use snake_case
type RoomInfo struct {
    PlayerCount int `json:"player_count"`
}
```

### TypeScript Frontend

```typescript
// Use composition API with <script setup>
<script setup lang="ts">
import { ref, computed } from 'vue';

// Type imports
import type { Game, User } from '../types';

// Store usage
const store = useGameStore();
const { t } = useLocaleStore();
</script>

// CSS use CSS variables for theming
<style scoped>
.element {
  background: var(--bg-primary);
  color: var(--text-primary);
}
</style>
```

### CSS Variables

```css
/* Available theme variables */
--bg-primary
--bg-secondary
--bg-tertiary
--bg-quaternary
--text-primary
--text-secondary
--text-muted
--border-color
--border-light
--accent-color
--accent-hover
--success-color
--warning-color
--red-color
--blue-color
--text-on-accent
```

## Adding New Features

### Adding a New Message Type

1. **Backend** (`pkg/protocol/protocol.go`):
```go
const (
    MessageTypeNewFeature = "new_feature"
)

type NewFeaturePayload struct {
    // fields
}
```

2. **Backend Handler** (`internal/websocket/handler.go`):
```go
case protocol.MessageTypeNewFeature:
    // handle message
```

3. **Frontend Types** (`frontend/src/types/index.ts`):
```typescript
export type MessageType = 
  | 'new_feature'
  // ...

export interface NewFeaturePayload {
  // fields
}
```

4. **Frontend WebSocket** (`frontend/src/composables/useWebSocket.ts`):
```typescript
function newFeature(params: NewFeaturePayload) {
  send({ type: 'new_feature', payload: params });
}

// Handle response in onMessage
```

### Adding a New Language

1. Create `frontend/src/locales/xx-XX.ts`
2. Add to `frontend/src/locales/index.ts`:
```typescript
import xxXX from './xx-XX';

export const locales = {
  'xx-XX': xxXX,
};

export const localeNames = {
  'xx-XX': 'Language Name',
};
```

### Adding a New Game Rule

1. **Backend** (`internal/game/types.go`):
```go
type GameRule string

const (
    RuleNewRule GameRule = "new_rule"
)
```

2. **Backend** (`internal/game/game.go`):
```go
case RuleNewRule:
    if err := g.markNewRule(cell, player); err != nil {
        return err
    }
```

3. **Frontend Types** (`frontend/src/types/index.ts`):
```typescript
export type GameRule = 'normal' | 'blackout' | 'phase' | 'new_rule';
```

4. **Frontend Locale** (both language files):
```typescript
rule: {
  newRule: 'New Rule Name',
}
```

## Development Commands

```bash
# Frontend development
cd frontend
npm install
npm run dev

# Build frontend
npm run build

# Run Wails development
wails dev

# Build Wails application
wails build

# Run standalone server
go run ./cmd/server

# Run standalone server with custom options
go run ./cmd/server --port 8765 --data ./data --room-ttl 30m
```

## WebSocket Protocol

### Client → Server Messages

| Type | Description | Payload |
|------|-------------|---------|
| `set_name` | Set player name | `{ name: string }` |
| `create_room` | Create new room | `{ name, password }` |
| `join_room` | Join room | `{ room_id, password }` |
| `leave_room` | Leave current room | - |
| `list_rooms` | Request room list | - |
| `set_role` | Set user role | `{ user_id, role, color }` |
| `start_game` | Start the game | - |
| `reset_game` | Reset the game | - |
| `mark_cell` | Mark a cell | `{ row, col, color }` |
| `unmark_cell` | Unmark a cell | `{ row, col }` |
| `set_cell_text` | Set cell text | `{ row, col, text }` |
| `set_rule` | Set game rule | `{ rule }` |
| `set_password` | Set room password | `{ password }` |
| `clear_cell_mark` | Clear specific mark color | `{ row, col, color }` |
| `settle` | Settle game (phase rule) | `{ player }` |

### Server → Client Messages

| Type | Description | Payload |
|------|-------------|---------|
| `state_update` | Full state sync | `StateUpdate` |
| `room_list` | Room list response | `RoomInfo[]` |
| `error` | Error message | `{ code, message }` |
| `connected` | Connection confirmed | `{ user_id }` |

## State Management

### Game Store (`stores/game.ts`)

```typescript
// State
connected: boolean
userId: string
userName: string
currentRoom: Room | null
game: Game | null
users: User[]
roomList: RoomInfo[]
error: string | null

// Getters
currentUser: User | undefined
isOwner: boolean
isReferee: boolean
isPlayer: boolean
inRoom: boolean
redPlayer: User | undefined
bluePlayer: User | undefined
spectators: User[]
```

## Important Notes

1. **Thread Safety**: Backend Room uses `sync.RWMutex` for concurrent access
2. **State Sync**: Full state is sent on every change via `state_update`
3. **Owner Transfer**: When owner leaves, ownership transfers to next user
4. **Cell Text**: Supports newlines via `\n` character
5. **Auto Font Size**: Cell text auto-sizes based on content length
6. **LocalStorage Keys**:
   - `bingosync-theme`
   - `bingosync-locale`
   - `bingosync-player-name`
   - `bingosync-server-url`
   - `bingosync-streamer-mode`

