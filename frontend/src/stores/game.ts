import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { Game, Room, User, RoomInfo, StateUpdate } from '../types';

export const useGameStore = defineStore('game', () => {
  // State
  const connected = ref(false);
  const userId = ref<string>('');
  const userName = ref<string>('');
  const currentRoom = ref<Room | null>(null);
  const game = ref<Game | null>(null);
  const users = ref<User[]>([]);
  const roomList = ref<RoomInfo[]>([]);
  const error = ref<string | null>(null);
  const streamToken = ref<string | null>(null);

  // WebSocket state (shared across components)
  const ws = ref<WebSocket | null>(null);
  const connecting = ref(false);

  // Getters
  const currentUser = computed(() => users.value.find(u => u.id === userId.value));
  const isOwner = computed(() => currentRoom.value?.owner_id === userId.value);
  const isReferee = computed(() => currentUser.value?.role === 'referee');
  const isPlayer = computed(() => currentUser.value?.role === 'player');
  const isSpectator = computed(() => currentUser.value?.role === 'spectator');
  const inRoom = computed(() => currentRoom.value !== null);
  
  const redPlayer = computed(() => users.value.find(u => u.player_color === 'red'));
  const bluePlayer = computed(() => users.value.find(u => u.player_color === 'blue'));
  const spectators = computed(() => users.value.filter(u => u.role === 'spectator'));

  // Actions
  function setConnected(value: boolean) {
    connected.value = value;
  }

  function setUserInfo(id: string, name: string) {
    userId.value = id;
    userName.value = name;
  }

  function setStateUpdate(data: StateUpdate) {
    currentRoom.value = data.room;
    game.value = data.game;
    users.value = data.users;
  }

  function setRoomList(rooms: RoomInfo[]) {
    roomList.value = rooms;
  }

  function leaveRoom() {
    currentRoom.value = null;
    game.value = null;
    users.value = [];
    streamToken.value = null;
  }

  function setStreamToken(token: string) {
    streamToken.value = token;
  }

  function reset() {
    connected.value = false;
    userId.value = '';
    userName.value = '';
    currentRoom.value = null;
    game.value = null;
    users.value = [];
    roomList.value = [];
    error.value = null;
    streamToken.value = null;
    ws.value = null;
    connecting.value = false;
  }

  function setError(msg: string | null) {
    error.value = msg;
  }

  function clearError() {
    error.value = null;
  }

  return {
    // State
    connected,
    userId,
    userName,
    currentRoom,
    game,
    users,
    roomList,
    error,
    streamToken,
    ws,
    connecting,
    // Getters
    currentUser,
    isOwner,
    isReferee,
    isPlayer,
    isSpectator,
    inRoom,
    redPlayer,
    bluePlayer,
    spectators,
    // Actions
    setConnected,
    setUserInfo,
    setStateUpdate,
    setRoomList,
    leaveRoom,
    setStreamToken,
    reset,
    setError,
    clearError,
  };
});
