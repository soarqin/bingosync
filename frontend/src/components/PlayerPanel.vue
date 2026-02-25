<template>
  <div class="player-panel">
    <h3>{{ t('player.players') }}</h3>
    
    <div class="player-section">
      <div class="player red">
        <span class="color-dot red"></span>
        <span class="name">
          {{ redPlayer?.name || t('player.unassigned') }}
          <span v-if="redPlayer && isRoomOwner(redPlayer.id)" class="owner-tag">{{ t('room.owner') }}</span>
        </span>
        <template v-if="redPlayer">
          <button v-if="redPlayer.id === currentUser?.id" @click="becomeSpectator">{{ t('common.cancel') }}</button>
          <button v-else-if="isOwner" @click="removeRole(redPlayer.id)">{{ t('player.remove') }}</button>
        </template>
        <button v-else-if="canAssign" @click="becomePlayer('red')">{{ t('common.save') }}</button>
      </div>
      
      <div class="player blue">
        <span class="color-dot blue"></span>
        <span class="name">
          {{ bluePlayer?.name || t('player.unassigned') }}
          <span v-if="bluePlayer && isRoomOwner(bluePlayer.id)" class="owner-tag">{{ t('room.owner') }}</span>
        </span>
        <template v-if="bluePlayer">
          <button v-if="bluePlayer.id === currentUser?.id" @click="becomeSpectator">{{ t('common.cancel') }}</button>
          <button v-else-if="isOwner" @click="removeRole(bluePlayer.id)">{{ t('player.remove') }}</button>
        </template>
        <button v-else-if="canAssign" @click="becomePlayer('blue')">{{ t('common.save') }}</button>
      </div>
    </div>

    <div class="referee-section">
      <h4>{{ t('player.referee') }}</h4>
      <div class="referee">
        <span class="name">
          {{ referee?.name || t('player.unassigned') }}
          <span v-if="referee && isRoomOwner(referee.id)" class="owner-tag">{{ t('room.owner') }}</span>
        </span>
        <template v-if="referee">
          <button v-if="referee.id === currentUser?.id" @click="becomeSpectator">{{ t('common.cancel') }}</button>
          <button v-else-if="isOwner" @click="removeRole(referee.id)">{{ t('player.remove') }}</button>
        </template>
        <button v-else-if="canAssign" @click="becomeReferee">{{ t('common.save') }}</button>
      </div>
    </div>

    <div class="spectators-section">
      <h4>{{ t('player.spectator') }} ({{ spectators.length }})</h4>
      <div v-for="user in spectators" :key="user.id" class="spectator">
        <span>
          {{ user.name }}
          <span v-if="isRoomOwner(user.id)" class="owner-tag">{{ t('room.owner') }}</span>
          <span v-if="user.id === currentUser?.id" class="you-tag">({{ t('connection.yourName').toLowerCase() }})</span>
        </span>
        <template v-if="isOwner && user.id !== currentUser?.id">
          <button @click="assignRoleToUser(user.id, 'player', 'red')">{{ t('game.red') }}</button>
          <button @click="assignRoleToUser(user.id, 'player', 'blue')">{{ t('game.blue') }}</button>
          <button @click="assignRoleToUser(user.id, 'referee')">{{ t('player.referee') }}</button>
        </template>
      </div>
    </div>

    <div class="current-user" v-if="currentUser">
      <span>{{ t('player.yourRole') }}: {{ roleText }}</span>
      <span v-if="isPlayer" :class="currentUser.player_color">
        ({{ currentUser.player_color === 'red' ? t('game.redTeam') : t('game.blueTeam') }})
      </span>
      <div class="role-actions">
        <button v-if="!isPlayer" @click="becomePlayer('red')">{{ t('player.becomeRed') }}</button>
        <button v-if="!isPlayer" @click="becomePlayer('blue')">{{ t('player.becomeBlue') }}</button>
        <button v-if="!isReferee" @click="becomeReferee">{{ t('player.becomeReferee') }}</button>
        <button v-if="isPlayer || isReferee" @click="becomeSpectator">{{ t('player.becomeSpectator') }}</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { Game } from '../types';
import { useGameStore } from '../stores/game';
import { useWebSocket } from '../composables/useWebSocket';
import { useLocaleStore } from '../stores/locale';

const props = defineProps<{
  game: Game | null;
}>();

const store = useGameStore();
const { setRole } = useWebSocket();
const { t } = useLocaleStore();

const currentUser = computed(() => store.currentUser);
const isOwner = computed(() => store.isOwner);
const isPlayer = computed(() => store.isPlayer);
const isReferee = computed(() => store.isReferee);
const redPlayer = computed(() => store.redPlayer);
const bluePlayer = computed(() => store.bluePlayer);
const spectators = computed(() => store.spectators);
const roomOwnerId = computed(() => store.currentRoom?.owner_id);

// Get referee
const referee = computed(() => store.users.find(u => u.role === 'referee'));

// Check if a user is the room owner
function isRoomOwner(userId: string): boolean {
  return roomOwnerId.value === userId;
}

// Whether user can choose their own role (any user in game can)
const canAssign = computed(() => !!currentUser.value);

const roleText = computed(() => {
  if (isReferee.value) return t('player.referee');
  if (isPlayer.value) return t('player.player');
  return t('player.spectator');
});

// Become player
function becomePlayer(color: 'red' | 'blue') {
  if (currentUser.value) {
    setRole(currentUser.value.id, 'player', color);
  }
}

// Become referee
function becomeReferee() {
  if (currentUser.value) {
    setRole(currentUser.value.id, 'referee');
  }
}

// Become spectator
function becomeSpectator() {
  if (currentUser.value) {
    setRole(currentUser.value.id, 'spectator');
  }
}

function assignRoleToUser(userId: string, role: string, color?: string) {
  setRole(userId, role, color);
}

function removeRole(userId: string) {
  setRole(userId, 'spectator');
}
</script>

<style scoped>
.player-panel {
  padding: 20px;
  background: var(--bg-tertiary);
  border-radius: 8px;
}

h3 {
  margin: 0 0 15px 0;
  color: var(--text-primary);
}

.player-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 20px;
}

.player {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  background: var(--bg-primary);
  border-radius: 4px;
  border-left: 4px solid transparent;
}

.player.red {
  border-left-color: var(--red-color);
}

.player.blue {
  border-left-color: var(--blue-color);
}

.color-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.color-dot.red {
  background: var(--red-color);
}

.color-dot.blue {
  background: var(--blue-color);
}

.name {
  flex: 1;
  color: var(--text-primary);
}

.referee-section {
  margin-bottom: 20px;
}

.referee-section h4 {
  margin: 0 0 10px 0;
  font-size: 14px;
  color: var(--text-muted);
}

.referee {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  background: var(--bg-primary);
  border-radius: 4px;
  border-left: 4px solid var(--warning-color);
}

.referee .name {
  color: var(--text-primary);
}

button {
  padding: 4px 8px;
  font-size: 12px;
}

.spectators-section h4 {
  margin: 0 0 10px 0;
  font-size: 14px;
  color: var(--text-muted);
}

.spectator {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 5px 0;
  font-size: 14px;
  color: var(--text-secondary);
}

.spectator span {
  flex: 0;
}

.spectator span:first-child {
  flex: 1;
}

.you-tag {
  color: var(--accent-color);
  font-size: 12px;
}

.owner-tag {
  display: inline-block;
  background: linear-gradient(135deg, #f39c12, #e67e22);
  color: #fff;
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 3px;
  margin-left: 6px;
  font-weight: bold;
  vertical-align: middle;
}

.current-user {
  margin-top: 20px;
  padding-top: 15px;
  border-top: 1px solid var(--border-light);
  font-weight: bold;
  color: var(--text-primary);
}

.role-actions {
  display: flex;
  gap: 8px;
  margin-top: 10px;
  flex-wrap: wrap;
}

.role-actions button {
  padding: 6px 12px;
  font-size: 12px;
}

.red {
  color: var(--red-color);
}

.blue {
  color: var(--blue-color);
}
</style>
