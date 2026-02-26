<script lang="ts" setup>
import { ref, onMounted, computed, watch, provide } from 'vue';
import BingoBoard from './components/BingoBoard.vue';
import RoomList from './components/RoomList.vue';
import RoomSettings from './components/RoomSettings.vue';
import PlayerPanel from './components/PlayerPanel.vue';
import type { PlayerColor } from './types';
import { useGameStore } from './stores/game';
import { useWebSocket } from './composables/useWebSocket';
import { useThemeStore } from './stores/theme';
import { useLocaleStore } from './stores/locale';
import type { LocaleCode } from './locales';

const store = useGameStore();
const themeStore = useThemeStore();
const localeStore = useLocaleStore();

const { 
  connect, 
  disconnect, 
  listRooms,
  startGame,
  resetGame,
  markCell,
  unmarkCell,
  leaveRoom,
  setName,
  settle
} = useWebSocket();

const { t } = localeStore;

const serverUrl = ref('ws://localhost:8765/ws');
const playerName = ref('');
const selectedColor = ref<PlayerColor>('none');
const bingoBoardRef = ref<InstanceType<typeof BingoBoard> | null>(null);

const connected = computed(() => store.connected);
const connecting = computed(() => store.connecting);
const inRoom = computed(() => store.inRoom);
const game = computed(() => store.game);
const currentRoom = computed(() => store.currentRoom);

// Settlement related computed properties
const isCurrentPlayerSettled = computed(() => {
  if (!game.value) return false;
  const color = store.currentUser?.player_color;
  if (color === 'red') return game.value.red_settled ?? false;
  if (color === 'blue') return game.value.blue_settled ?? false;
  return false;
});

const canSettleNow = computed(() => {
  if (!game.value) return false;
  
  // If someone already settled first, second player can settle without conditions
  if (game.value.first_settler && game.value.first_settler !== 'none') {
    return true;
  }
  
  // First settler must meet conditions
  const color = store.currentUser?.player_color;
  if (color === 'red') {
    return (game.value.red_row_marks?.[4] ?? 0) >= 2;
  }
  if (color === 'blue') {
    return (game.value.blue_row_marks?.[4] ?? 0) >= 2;
  }
  return false;
});

const canRedSettle = computed(() => {
  if (!game.value) return false;
  
  // If someone already settled first, red can settle without conditions
  if (game.value.first_settler && game.value.first_settler !== 'none') {
    return true;
  }
  
  // Otherwise must meet conditions
  return (game.value.red_row_marks?.[4] ?? 0) >= 2;
});

const canBlueSettle = computed(() => {
  if (!game.value) return false;
  
  // If someone already settled first, blue can settle without conditions
  if (game.value.first_settler && game.value.first_settler !== 'none') {
    return true;
  }
  
  // Otherwise must meet conditions
  return (game.value.blue_row_marks?.[4] ?? 0) >= 2;
});

// Provide selectedColor to child components (BingoBoard)
provide('selectedColor', selectedColor);

async function handleConnect() {
  if (connected.value) {
    disconnect();
  } else {
    try {
      await connect(serverUrl.value);
      // Save server address after successful connection
      localStorage.setItem('bingosync-server-url', serverUrl.value);
      // Set username after successful connection
      if (playerName.value.trim()) {
        const name = playerName.value.trim();
        setName(name);
        localStorage.setItem('bingosync-player-name', name);
      }
      listRooms();
    } catch (e) {
      console.error('Connection failed:', e);
      store.setError(t('connection.connectionFailed'));
    }
  }
}

function handleSaveName() {
  if (connected.value && !inRoom.value && playerName.value.trim()) {
    const name = playerName.value.trim();
    setName(name);
    localStorage.setItem('bingosync-player-name', name);
  }
}

function handleLeaveRoom() {
  leaveRoom();
}

// File input ref for import
const fileInputRef = ref<HTMLInputElement | null>(null);

function handleImportClick() {
  fileInputRef.value?.click();
}

async function handleFileImport(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0];
  if (!file) return;
  
  try {
    const text = await file.text();
    let texts: string[] = [];
    
    if (file.name.toLowerCase().endsWith('.txt')) {
      // TXT: one per line, 25 lines total
      const lines = text.split('\n').map(line => {
        return line.trim().replace(/\\n/g, '\n');
      }).filter(line => line.length > 0);
      texts = lines.slice(0, 25);
    } else if (file.name.toLowerCase().endsWith('.csv')) {
      // CSV: 5 per line, 5 lines total
      const lines = text.split('\n').slice(0, 5);
      for (const line of lines) {
        const cols = parseCSVLine(line);
        texts.push(...cols.slice(0, 5));
      }
    }
    
    // Ensure 25 elements
    while (texts.length < 25) {
      texts.push('');
    }
    texts = texts.slice(0, 25);
    
    // Use WebSocket to set all cell texts
    const { setAllCellTexts } = useWebSocket();
    setAllCellTexts(texts);
  } catch (e) {
    console.error('Failed to import file:', e);
    store.setError(t('settings.importFailed'));
  }
  
  // Reset file input
  (event.target as HTMLInputElement).value = '';
}

function parseCSVLine(line: string): string[] {
  const result: string[] = [];
  let current = '';
  let inQuotes = false;
  
  for (let i = 0; i < line.length; i++) {
    const char = line[i];
    
    if (char === '"') {
      if (inQuotes && line[i + 1] === '"') {
        current += '"';
        i++;
      } else {
        inQuotes = !inQuotes;
      }
    } else if (char === ',' && !inQuotes) {
      result.push(current.trim());
      current = '';
    } else {
      current += char;
    }
  }
  
  result.push(current.trim());
  return result;
}

function handleExport() {
  if (!game.value?.board) return;
  
  // Collect all cell texts
  const texts: string[] = [];
  for (const row of game.value.board.cells) {
    for (const cell of row) {
      texts.push(cell.text || '');
    }
  }
  
  // Generate CSV content
  const csvLines: string[] = [];
  for (let i = 0; i < 5; i++) {
    const rowTexts = texts.slice(i * 5, (i + 1) * 5);
    const csvRow = rowTexts.map(text => {
      if (text.includes(',') || text.includes('"') || text.includes('\n')) {
        return '"' + text.replace(/"/g, '""') + '"';
      }
      return text;
    }).join(',');
    csvLines.push(csvRow);
  }
  
  const csvContent = csvLines.join('\n');
  
  // Download file
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = 'bingo-board.csv';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

function handleMark(row: number, col: number, color: PlayerColor) {
  if (color === 'none') {
    // Unmark cell
    unmarkCell(row, col);
  } else {
    // Mark cell
    markCell(row, col, color);
  }
}

function handleSettle(color: PlayerColor) {
  settle(color);
}

function handlePlayerSettle() {
  const color = store.currentUser?.player_color;
  if (color && color !== 'none') {
    settle(color);
  }
}

// Watch for username changes in store, sync to input
watch(() => store.userName, (newName) => {
  if (newName && !playerName.value) {
    playerName.value = newName;
  }
});

// Save to localStorage when locale changes
function onLocaleChange(event: Event) {
  const code = (event.target as HTMLSelectElement).value as LocaleCode;
  localeStore.setLocale(code);
}

onMounted(() => {
  // Load user preferences
  themeStore.loadTheme();
  localeStore.loadLocale();
  
  // Load player name
  const savedName = localStorage.getItem('bingosync-player-name');
  if (savedName) {
    playerName.value = savedName;
  }
  
  // Load server address (prefer localStorage, fallback to Wails backend)
  const savedServerUrl = localStorage.getItem('bingosync-server-url');
  if (savedServerUrl) {
    serverUrl.value = savedServerUrl;
  } else if (window.go && window.go.main && window.go.main.App) {
    window.go.main.App.GetServerURL().then((url: string) => {
      serverUrl.value = url;
    }).catch(() => {});
  }
});
</script>

<template>
  <div id="app">
    <!-- Header -->
    <header class="header">
      <div class="header-left">
        <h1>BingoSync</h1>
      </div>
      <div class="header-center">
        <div class="connection">
          <div class="name-wrapper">
            <input 
              v-model="playerName" 
              :placeholder="t('connection.yourName')" 
              :disabled="inRoom"
              class="name-input"
              @keyup.enter="handleSaveName"
            />
            <button 
              v-if="connected && !inRoom && playerName.trim()" 
              @click="handleSaveName"
              class="save-name-btn"
            >{{ t('common.save') }}</button>
          </div>
          <input v-model="serverUrl" :placeholder="t('connection.serverAddress')" :disabled="connected" />
          <button @click="handleConnect" :disabled="connecting">
            {{ connecting ? t('common.connecting') : connected ? t('common.disconnect') : t('common.connect') }}
          </button>
        </div>
      </div>
      <div class="header-right">
        <select 
          v-model="localeStore.locale" 
          class="locale-select"
          @change="onLocaleChange"
        >
          <option 
            v-for="loc in localeStore.availableLocales" 
            :key="loc.code" 
            :value="loc.code"
          >
            {{ loc.name }}
          </option>
        </select>
        <button @click="themeStore.toggleTheme" class="icon-btn theme-btn">
          {{ themeStore.theme === 'dark' ? '🌙' : '☀️' }}
        </button>
      </div>
    </header>

    <!-- Error display at bottom -->
    <div v-if="store.error" class="error-bar">
      {{ store.error }}
      <button @click="store.clearError()">{{ t('common.close') }}</button>
    </div>

    <!-- Main content -->
    <main class="main">
      <!-- Lobby view -->
      <template v-if="!inRoom">
        <RoomList v-if="connected" />
        <div v-else class="disconnected">
          <p>{{ t('connection.pleaseConnect') }}</p>
        </div>
      </template>

      <!-- Game room view -->
      <template v-else>
        <div class="room-header">
          <h2>{{ currentRoom?.name }}</h2>
          
          <!-- Board controls in header center -->
          <div class="board-controls-header">
            <!-- Import/Export buttons (owner only, waiting status) -->
            <template v-if="store.isOwner && game?.status === 'waiting'">
              <input 
                ref="fileInputRef"
                type="file" 
                accept=".txt,.csv" 
                @change="handleFileImport"
                style="display: none"
              />
              <button @click="handleImportClick" class="control-btn import-btn">
                📥 {{ t('game.importText') }}
              </button>
              <button @click="handleExport" class="control-btn export-btn">
                📤 {{ t('game.exportText') }}
              </button>
            </template>
            
            <!-- Start/Reset button (owner only) -->
            <button 
              v-if="store.isOwner"
              @click="game?.status === 'waiting' ? startGame() : resetGame()" 
              class="control-btn"
              :class="game?.status === 'waiting' ? 'start-btn' : 'reset-btn'"
            >
              <template v-if="game?.status === 'waiting'">🎮 {{ t('game.startGame') }}</template>
              <template v-else>🔄 {{ game?.status === 'finished' ? t('game.restart') : t('game.resetBoard') }}</template>
            </button>
          </div>
          
          <div class="room-actions">
            <!-- Room settings button -->
            <RoomSettings :game="game" />
            <button @click="handleLeaveRoom" class="leave-btn">{{ t('room.leaveRoom') }}</button>
          </div>
        </div>
        
        <div class="game-container">
          <div class="left-panel">
            <BingoBoard
              v-if="game?.board" 
              ref="bingoBoardRef"
              :board="game.board" 
              :game="game"
              @mark="handleMark"
              @settle="handleSettle"
            />
          </div>
          
          <div class="right-panel">
            <PlayerPanel :game="game" />
            
            <!-- Control buttons section -->
            <div class="control-section">
              <!-- Referee color picker -->
              <div v-if="store.isReferee && (game?.status === 'playing' || game?.status === 'finished')" class="color-picker">
                <span>{{ t('game.selectColor') }}:</span>
                <div class="color-buttons">
                  <button 
                    class="color-btn red" 
                    :class="{ active: selectedColor === 'red' }"
                    @click="selectedColor = 'red'"
                  >{{ t('game.red') }}</button>
                  <button 
                    class="color-btn blue" 
                    :class="{ active: selectedColor === 'blue' }"
                    @click="selectedColor = 'blue'"
                  >{{ t('game.blue') }}</button>
                  <button 
                    class="color-btn clear" 
                    :class="{ active: selectedColor === 'none' }"
                    @click="selectedColor = 'none'"
                  >{{ t('game.clear') }}</button>
                </div>
              </div>
              
              <!-- Settlement buttons for phase rule -->
              <template v-if="game?.rule === 'phase' && game?.status === 'playing'">
                <!-- For players -->
                <template v-if="store.isPlayer">
                  <button 
                    v-if="!isCurrentPlayerSettled"
                    @click="handlePlayerSettle" 
                    class="control-btn settle-btn"
                    :disabled="!canSettleNow"
                  >
                    ⚖️ {{ t('phase.settle') }}
                  </button>
                  <div v-else class="settled-status">
                    ✓ {{ t('phase.settled') }}
                  </div>
                </template>
                <!-- For referee -->
                <template v-else-if="store.isReferee">
                  <div class="settle-buttons">
                    <button 
                      @click="handleSettle('red')" 
                      class="control-btn settle-btn red"
                      :disabled="game?.red_settled || !canRedSettle"
                    >
                      ⚖️ {{ t('game.redTeam') }}
                    </button>
                    <button 
                      @click="handleSettle('blue')" 
                      class="control-btn settle-btn blue"
                      :disabled="game?.blue_settled || !canBlueSettle"
                    >
                      ⚖️ {{ t('game.blueTeam') }}
                    </button>
                  </div>
                </template>
              </template>

              <!-- Waiting state message -->
              <div v-if="game?.status === 'waiting'" class="waiting-status">
                <span class="waiting-icon">⏳</span>
                <span class="waiting-text">{{ t('game.waiting') }}</span>
              </div>
            </div>
          </div>
        </div>
      </template>
    </main>
  </div>
</template>

<style>
/* CSS Variables for theming */
:root {
  --bg-primary: #1a1a2e;
  --bg-secondary: #16213e;
  --bg-tertiary: #2a2a4a;
  --bg-quaternary: #333;
  --text-primary: #eee;
  --text-secondary: #ccc;
  --text-muted: #888;
  --border-color: #0f3460;
  --border-light: #3a3a5a;
  --accent-color: #e94560;
  --accent-hover: #ff6b6b;
  --success-color: #27ae60;
  --success-hover: #2ecc71;
  --info-color: #3498db;
  --info-hover: #2980b9;
  --warning-color: #f39c12;
  --disabled-color: #4a4a6a;
  --red-color: #e74c3c;
  --blue-color: #3498db;
  --cell-bg: #fff;
  --cell-text: #333;
  --cell-empty-bg: #ecf0f1;
  --cell-times-color: #666;
  --text-on-accent: #fff;
  --error-bg: #e94560;
  --clear-btn-bg: #95a5a6;
}

:root.light {
  --bg-primary: #f5f5f5;
  --bg-secondary: #ffffff;
  --bg-tertiary: #e8e8e8;
  --bg-quaternary: #ddd;
  --text-primary: #333;
  --text-secondary: #555;
  --text-muted: #888;
  --border-color: #ccc;
  --border-light: #ddd;
  --accent-color: #e94560;
  --accent-hover: #d63050;
  --success-color: #27ae60;
  --success-hover: #1e8449;
  --info-color: #3498db;
  --info-hover: #2980b9;
  --warning-color: #f39c12;
  --disabled-color: #bbb;
  --red-color: #e74c3c;
  --blue-color: #3498db;
  --cell-bg: #fff;
  --cell-text: #333;
  --cell-empty-bg: #f0f0f0;
  --cell-times-color: #666;
  --text-on-accent: #fff;
  --error-bg: #e94560;
  --clear-btn-bg: #95a5a6;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
  background: var(--bg-primary);
  color: var(--text-primary);
  min-height: 100vh;
}

#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  gap: 20px;
}

.header-left {
  flex-shrink: 0;
}

.header h1 {
  font-size: 24px;
  color: var(--accent-color);
  white-space: nowrap;
}

.header-center {
  flex: 1;
  display: flex;
  justify-content: center;
}

.header-right {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.connection {
  display: flex;
  gap: 10px;
  align-items: center;
}

.name-wrapper {
  display: flex;
  gap: 5px;
  align-items: center;
}

.connection input {
  width: 300px;
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background: var(--bg-primary);
  color: var(--text-primary);
}

.connection .name-input {
  width: 120px;
}

.save-name-btn {
  padding: 8px 12px;
  font-size: 12px;
  background: var(--success-color);
  white-space: nowrap;
}

.save-name-btn:hover {
  background: var(--success-hover);
}

.icon-btn {
  padding: 8px 12px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  min-width: 50px;
}

.icon-btn:hover {
  background: var(--bg-quaternary);
}

.locale-select {
  padding: 8px 12px;
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: bold;
  min-width: 70px;
}

.locale-select:hover {
  background: var(--bg-quaternary);
}

.locale-select:focus {
  outline: none;
  border-color: var(--accent-color);
}

.theme-btn {
  font-size: 16px;
}

.error-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background: var(--error-bg);
  color: var(--text-on-accent);
  z-index: 1000;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.3);
}

.error-bar button {
  background: rgba(255, 255, 255, 0.2);
  padding: 6px 12px;
}

.error-bar button:hover {
  background: rgba(255, 255, 255, 0.3);
}

.main {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
  overflow-x: hidden;
}

.disconnected {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  font-size: 18px;
  color: var(--text-muted);
}

.room-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  position: relative;
}

.room-header h2 {
  font-size: 20px;
}

.room-actions {
  display: flex;
  gap: 10px;
  align-items: center;
}

.room-actions button {
  padding: 8px 16px;
}

.leave-btn {
  background: var(--accent-color);
}

.leave-btn:hover {
  background: var(--accent-hover);
}

.game-container {
  display: flex;
  gap: 20px;
  overflow: hidden;
  width: 100%;
  box-sizing: border-box;
}

.left-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-width: 0;
  overflow: hidden;
}

.board-controls-header {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  justify-content: center;
  gap: 10px;
}

.board-controls-header .control-btn {
  padding: 8px 16px;
  font-weight: 500;
  font-size: 13px;
}

.board-controls-header .import-btn {
  background: var(--info-color);
}

.board-controls-header .import-btn:hover {
  background: var(--info-hover);
}

.board-controls-header .export-btn {
  background: var(--success-color);
}

.board-controls-header .export-btn:hover {
  background: var(--success-hover);
}

.board-controls-header .start-btn {
  background: var(--success-color);
  font-weight: bold;
}

.board-controls-header .start-btn:hover {
  background: var(--success-hover);
}

.board-controls-header .reset-btn {
  background: var(--warning-color);
  font-weight: bold;
}

.board-controls-header .reset-btn:hover {
  background: #e67e22;
}

.right-panel {
  width: 300px;
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.control-section {
  padding: 15px;
  background: var(--bg-tertiary);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.waiting-status {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 20px 10px;
  color: var(--text-muted);
  font-size: 14px;
}

.waiting-icon {
  font-size: 20px;
  animation: pulse 2s infinite;
}

.waiting-text {
  font-weight: 500;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.control-btn {
  padding: 10px 16px;
  font-size: 14px;
  font-weight: bold;
}

.color-picker {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px;
  background: var(--bg-primary);
  border-radius: 6px;
}

.color-picker span {
  color: var(--text-secondary);
  font-size: 14px;
  font-weight: bold;
}

.color-buttons {
  display: flex;
  gap: 8px;
}

.color-btn {
  flex: 1;
  padding: 8px 12px;
  border: 2px solid transparent;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
  transition: all 0.2s;
  font-size: 13px;
}

.color-btn.red {
  background: var(--red-color);
  color: white;
}

.color-btn.blue {
  background: var(--blue-color);
  color: white;
}

.color-btn.clear {
  background: var(--clear-btn-bg);
  color: white;
}

.color-btn.active {
  border-color: var(--warning-color);
  transform: scale(1.05);
  box-shadow: 0 0 10px rgba(243, 156, 18, 0.5);
}

/* Settlement buttons */
.settle-btn {
  background: var(--info-color);
}

.settle-btn:hover:not(:disabled) {
  background: var(--info-hover);
}

.settle-btn.red {
  background: var(--red-color);
}

.settle-btn.red:hover:not(:disabled) {
  background: #c0392b;
}

.settle-btn.blue {
  background: var(--blue-color);
}

.settle-btn.blue:hover:not(:disabled) {
  background: #2980b9;
}

.settle-buttons {
  display: flex;
  gap: 8px;
}

.settle-buttons .settle-btn {
  flex: 1;
}

.settled-status {
  text-align: center;
  padding: 10px;
  background: var(--success-color);
  border-radius: 4px;
  color: white;
  font-weight: bold;
}

button {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  background: var(--accent-color);
  color: var(--text-on-accent);
  cursor: pointer;
  font-size: 14px;
}

button:hover:not(:disabled) {
  background: var(--accent-hover);
}

button:disabled {
  background: var(--disabled-color);
  cursor: not-allowed;
}

input, select {
  background: var(--bg-primary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
}

input:focus, select:focus {
  outline: none;
  border-color: var(--accent-color);
}
</style>

