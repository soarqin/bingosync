# BingoSync

Bingo board sync tool built with Go + Wails (Vue.js + TypeScript for frontend).

## Features

- **Real-time Multiplayer** - Play Bingo with friends in real-time
- **Multiple Game Rules** - Normal, Blackout, and Phase modes
- **Role System** - Player, Referee, and Spectator roles
- **Room Management** - Create rooms with optional password protection
- **Multi-language** - Supports Chinese (zh-CN) and English (en-US)
- **Theme Support** - Light and dark themes
- **Import/Export** - Import/export board text via CSV or TXT files

## Tech Stack

### Backend
- **Go 1.24**
- **Wails v2** - Desktop application framework
- **gws** - WebSocket library

### Frontend
- **Vue 3** + **TypeScript**
- **Pinia** - State management
- **Vite** - Build tool

## Installation

### Prerequisites

- Go 1.24 or later
- Node.js 18 or later
- Wails CLI (for desktop app)

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Clone and Run

```bash
# Clone the repository
git clone https://github.com/soarqin/bingosync.git
cd bingosync

# Install frontend dependencies
cd frontend
npm install
cd ..

# Run in development mode
wails dev
```

### Build Desktop Application

```bash
wails build
```

The built application will be in the `build/bin/` directory.

### Run Standalone Server

```bash
go run ./cmd/server --port 8765
```

Then connect to `ws://localhost:8765/ws` from the frontend.

## Usage

1. **Set your name** - Enter your player name before joining rooms
2. **Create or Join Room** - Create a new room or join an existing one
3. **Set Roles** - Room owner can assign players to Red/Blue teams and Referee
4. **Edit Board Text** - Room owner can customize the 5x5 Bingo board text
5. **Start Game** - Start the game when everyone is ready
6. **Mark Cells** - Players mark cells, Referee can mark/unmark any cell
7. **Win** - First to complete a line (Normal), full board (Blackout), or score-based (Phase)

## Game Rules

### Normal
- Each cell can only be marked once
- First player to complete 5 cells in a row (horizontal, vertical, or diagonal) wins
- If board is full without a line, player with more cells wins

### Blackout
- Players compete to mark all 25 cells
- First to complete the entire board wins

### Phase
- Rows unlock progressively
- Each row has a mark limit per player
- Score-based winning with phase configuration

## Development

### Project Structure

```
bingosync/
├── main.go              # Wails entry point
├── app.go               # Wails app bindings
├── cmd/server/          # Standalone server
├── internal/
│   ├── game/            # Game logic
│   ├── room/            # Room management
│   ├── user/            # User management
│   └── websocket/       # WebSocket handler
├── pkg/protocol/        # Message protocol
└── frontend/            # Vue.js frontend
```

### Commands

```bash
# Frontend development
cd frontend && npm run dev

# Build frontend
cd frontend && npm run build

# Run Wails dev
wails dev

# Build release
wails build

# Run standalone server
go run ./cmd/server
```

## License

MIT License

## Author

Soar Qin (soarchin@gmail.com)
