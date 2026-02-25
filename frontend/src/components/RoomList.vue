<template>
  <div class="room-list">
    <div class="header">
      <h2>{{ t('room.rooms') }}</h2>
      <button @click="refresh" :disabled="!connected">{{ t('room.refresh') }}</button>
    </div>
    
    <div class="create-room">
      <input v-model="newRoomName" :placeholder="t('room.roomName')" />
      <input v-model="newRoomPassword" type="password" :placeholder="t('room.password') + ' (' + t('room.optional') + ')'" />
      <button @click="handleCreate" :disabled="!connected || !newRoomName.trim()">
        {{ t('room.createRoom') }}
      </button>
    </div>

    <div class="rooms">
      <div v-if="rooms.length === 0" class="empty">
        {{ t('room.noRooms') }}
      </div>
      <div
        v-for="room in rooms"
        :key="room.id"
        class="room-item"
        @click="handleJoin(room)"
      >
        <div class="room-info">
          <span class="room-name">{{ room.name }}</span>
          <span class="room-meta">
            <span class="room-owner">{{ t('room.owner') }}: {{ room.owner_name || '-' }}</span>
            <span class="room-players">{{ room.player_count }} {{ t('room.people') }}</span>
          </span>
        </div>
        <span v-if="room.has_password" class="lock">🔒</span>
      </div>
    </div>

    <!-- Password dialog -->
    <div v-if="showPasswordDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>{{ t('room.password') }}</h3>
        <input v-model="joinPassword" type="password" :placeholder="t('room.password')" />
        <div class="dialog-buttons">
          <button @click="cancelJoin">{{ t('common.cancel') }}</button>
          <button @click="confirmJoin">{{ t('room.confirm') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import type { RoomInfo } from '../types';
import { useGameStore } from '../stores/game';
import { useWebSocket } from '../composables/useWebSocket';
import { useLocaleStore } from '../stores/locale';

const store = useGameStore();
const { listRooms, createRoom, joinRoom } = useWebSocket();
const { t } = useLocaleStore();

const newRoomName = ref('');
const newRoomPassword = ref('');
const showPasswordDialog = ref(false);
const joinPassword = ref('');
const selectedRoom = ref<RoomInfo | null>(null);

const connected = computed(() => store.connected);
const rooms = computed(() => store.roomList);

function refresh() {
  listRooms();
}

function handleCreate() {
  if (newRoomName.value.trim()) {
    createRoom(newRoomName.value.trim(), newRoomPassword.value || undefined);
    newRoomName.value = '';
    newRoomPassword.value = '';
  }
}

function handleJoin(room: RoomInfo) {
  selectedRoom.value = room;
  if (room.has_password) {
    showPasswordDialog.value = true;
    joinPassword.value = '';
  } else {
    joinRoom(room.id);
  }
}

function cancelJoin() {
  showPasswordDialog.value = false;
  selectedRoom.value = null;
  joinPassword.value = '';
}

function confirmJoin() {
  if (selectedRoom.value) {
    joinRoom(selectedRoom.value.id, joinPassword.value || undefined);
    cancelJoin();
  }
}

onMounted(() => {
  if (connected.value) {
    refresh();
  }
});
</script>

<style scoped>
.room-list {
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h2 {
  margin: 0;
}

.create-room {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.create-room input {
  flex: 1;
  padding: 10px;
  border: 1px solid var(--border-light);
  border-radius: 4px;
  background: var(--bg-primary);
  color: var(--text-primary);
}

.rooms {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.empty {
  text-align: center;
  color: var(--text-muted);
  padding: 40px;
}

.room-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
  background: var(--bg-tertiary);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.room-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
  background: var(--bg-quaternary);
}

.room-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.room-name {
  font-weight: bold;
  color: var(--text-primary);
}

.room-meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: var(--text-muted);
}

.room-owner {
  color: var(--text-secondary);
}

.room-players {
  color: var(--text-muted);
}

.lock {
  font-size: 20px;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
}

.dialog {
  background: var(--bg-tertiary);
  padding: 20px;
  border-radius: 8px;
  min-width: 300px;
}

.dialog h3 {
  margin: 0 0 15px 0;
  color: var(--text-primary);
}

.dialog input {
  width: 100%;
  padding: 10px;
  border: 1px solid var(--border-light);
  border-radius: 4px;
  margin-bottom: 15px;
  box-sizing: border-box;
  background: var(--bg-primary);
  color: var(--text-primary);
}

.dialog-buttons {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

button {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  background: var(--accent-color);
  color: var(--text-on-accent);
  cursor: pointer;
}

button:hover:not(:disabled) {
  background: var(--accent-hover);
}

button:disabled {
  background: var(--disabled-color);
  cursor: not-allowed;
}
</style>
