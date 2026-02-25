<script lang="ts" setup>
import { ref, onMounted, computed, watch } from 'vue';
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
  setName
} = useWebSocket();

const { t } = localeStore;

const serverUrl = ref('ws://localhost:8765/ws');
const playerName = ref('');

const connected = computed(() => store.connected);
const connecting = computed(() => store.connecting);
const inRoom = computed(() => store.inRoom);
const game = computed(() => store.game);
const currentRoom = computed(() => store.currentRoom);

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

function handleMark(row: number, col: number, color: PlayerColor) {
  if (color === 'none') {
    // Unmark cell
    unmarkCell(row, col);
  } else {
    // Mark cell
    markCell(row, col, color);
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

    <!-- Error display -->
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
          <div class="room-actions">
            <RoomSettings 
              :game="game"
              @reset="resetGame"
            />
            <button @click="handleLeaveRoom">{{ t('room.leaveRoom') }}</button>
          </div>
        </div>
        
        <div class="game-container">
          <div class="left-panel">
            <BingoBoard 
              v-if="game?.board" 
              :board="game.board" 
              :game="game"
              @mark="handleMark"
            />
          </div>
          
          <div class="right-panel">
            <PlayerPanel :game="game" />
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
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 20px;
  background: var(--error-bg);
  color: var(--text-on-accent);
}

.main {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
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

.game-container {
  display: flex;
  gap: 20px;
}

.left-panel {
  flex: 1;
  display: flex;
  justify-content: center;
}

.right-panel {
  width: 300px;
  display: flex;
  flex-direction: column;
  gap: 20px;
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

