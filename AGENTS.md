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
- **OBS Overlay**: SSE-based transparent overlay for OBS Browser Source (`/overlay?token=xxx&lang=xxx`)

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

### Protocol Version Policy

**IMPORTANT**: The protocol version must be incremented whenever making incompatible changes to the WebSocket protocol.

- Protocol version is defined in `pkg/protocol/protocol.go` as `ProtocolVersion` (integer)
- Frontend version is defined in `frontend/src/types/index.ts` as `PROTOCOL_VERSION`
- Server rejects connections with mismatched versions:
  - Client version < Server: returns 426 (client_outdated)
  - Client version > Server: returns 400 (server_outdated)
- **Always increment both versions together when making breaking changes**

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

1. **Thread Safety**: Backend Room uses `sync.RWMutex` for concurrent access. Lock ordering: `Manager.mu` must always be acquired before `Room.mu` to avoid deadlocks.
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
7. **OBS Overlay**: Stream tokens are bound to rooms (not users) and are persisted. Rebuilding the server binary re-embeds the overlay HTML — always rebuild after editing `internal/overlay/overlay.html`.
8. **SSE Subscribers**: Tracked in `Handler.sseSubscribers` (plain `map` protected by `sseSubMu RWMutex`). The `streamTokens` sync.Map is an in-memory index rebuilt from storage on startup.

## OBS Overlay Feature

### Architecture

The overlay uses Server-Sent Events (SSE) instead of WebSocket so OBS Browser Source needs zero interaction:

```
Wails Desktop  ──WebSocket──▶  Standalone Server (:8765)
                                  /ws     WebSocket
                                  /stream SSE push
                                  /overlay HTML page (embedded)

OBS Browser    ◀──SSE push────  /stream?token=xxx
Source          (passive, no sends)
```

### Token Lifecycle

- Token is generated on first `create_stream_token` message for a room
- Token is bound to the **room**, persisted to storage (Badger)
- Survives server restarts and user reconnections
- The same token is returned on subsequent requests for the same room
- Token is registered in `Handler.streamTokens` (sync.Map) on startup from persisted data

### New WebSocket Message Types

| Type | Direction | Description |
|------|-----------|-------------|
| `create_stream_token` | Client → Server | Request/retrieve the room's stream token |
| `stream_token` | Server → Client | Response with `{ token: string }` |

### New HTTP Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /overlay?token=xxx&lang=xxx` | Serves the OBS overlay HTML page |
| `GET /stream?token=xxx` | SSE endpoint streaming `state_update` events |

### Overlay Localisation

The overlay HTML (`internal/overlay/overlay.html`) supports `?lang=zh-CN` and `?lang=en-US`.
The frontend automatically appends the current UI language when copying the OBS URL.
Supported locale keys in `LOCALES`: `redTeam`, `blueTeam`, `waiting`, `playing`, `finished`, `wins`, `draw`, `connecting`, `notInRoom`, `noToken`.

## Recent Changes (2026-03-14)

### New Features
- **OBS Overlay via SSE**: Added `/overlay` and `/stream` HTTP endpoints to the standalone server
- **Stream Token**: New `create_stream_token` / `stream_token` WebSocket message pair; token is room-scoped and persisted (survives server restarts)
- **Overlay Localisation**: Overlay page supports `?lang=zh-CN` / `?lang=en-US`; frontend copies URL with current locale
- **Copy OBS URL button**: Added to room action bar; generates/reuses the room's stream token and copies the full overlay URL

### Bug Fixes
- **Race condition**: `broadcastRoomState` now iterates `state.Users` (snapshot from `GetState()`) instead of `r.Users` directly, eliminating a data race with concurrent room mutations
- **Double `store.Close()`**: Removed redundant `store.Close()` call in shutdown goroutine that could panic Badger
- **Missing broadcast on `create_room`**: When a user leaves an old room to create a new one, the old room's members now receive a `state_update`
- **Silent unmarshal failure**: `broadcastRoomState` now logs and skips on `json.Unmarshal` error instead of sending a zero-value state

### Refactoring
- **SSE subscriber registry**: Replaced redundant `sync.Map` + `RWMutex` combination with a plain `map` protected solely by `sseSubMu`
- **Protocol constants**: Added `MsgConnected` and `MsgNameSet` constants; removed unused `JoinedPayload`, `LeftPayload`, and `StreamTokenPayload.URL` field
- **Dead code removal** (Go): Removed `Handler.Log`, `Handler.GetRoomManager`, `Room.CancelDeletionTimer`, `ErrRoomFull`, `ErrWrongPassword`, `Manager.SetUserRole`, `Manager.SetUserPlayerColor`, `Manager.SetUserRoom`
- **Dead code removal** (Vue): Removed `handleFileImport`/`handleExport`/`parseCSVLine` from `BingoBoard.vue` (only live in `App.vue`); removed settlement `computed` props and `defineExpose` from `BingoBoard.vue`; removed unused `game` prop from `PlayerPanel.vue`; removed `bingoBoardRef`
- **App.vue**: Merged two `onMounted` blocks into one; replaced `460px` magic number with `--right-panel-width` CSS variable; fixed `::v-deep` → `:deep()` syntax; replaced polling loop in `handleCopyObsUrl` with a reactive `watch`
- **Storage GC**: Added periodic Badger value-log GC (`runGC`) to prevent unbounded disk growth
- **`player.you` locale key**: Added dedicated `you` / `你` translation key; replaced misuse of `connection.yourName` in `PlayerPanel`
- **IE fallback removed**: Removed `(navigator as any).userLanguage` cast in `locale.ts`

