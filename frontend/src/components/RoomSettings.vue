<template>
  <!-- 设置按钮 -->
  <button class="settings-btn" @click="showDialog = true">
    ⚙️ {{ t('common.settings') }}
  </button>

  <!-- 弹出对话框 -->
  <div v-if="showDialog" class="dialog-overlay" @click.self="showDialog = false">
    <div class="dialog-content">
      <div class="dialog-header">
        <h3>{{ t('common.settings') }}</h3>
        <button class="close-btn" @click="showDialog = false">✕</button>
      </div>
      
      <div class="dialog-body">
        <div class="setting-group">
          <label>{{ t('rule.gameRule') }}</label>
          <select v-model="selectedRule" :disabled="!canChangeSettings">
            <option value="normal">{{ t('rule.normal') }}</option>
            <option value="blackout">{{ t('rule.blackout') }}</option>
            <option value="phase">{{ t('rule.phase') }}</option>
          </select>
        </div>

        <template v-if="selectedRule === 'phase'">
          <div class="setting-group">
            <label>{{ t('settings.phaseConfig.rowScores') }}</label>
            <div class="row-scores">
              <input v-for="i in 5" :key="i" type="number" v-model.number="phaseConfig.row_scores[i-1]" min="0" />
            </div>
          </div>
          
          <div class="setting-group">
            <label>{{ t('settings.phaseConfig.secondHalfRate') }}</label>
            <input type="number" v-model.number="phaseConfig.second_half_rate" min="0" max="1" step="0.1" />
          </div>
          
          <div class="setting-group">
            <label>{{ t('settings.phaseConfig.finalBonus') }}</label>
            <input type="number" v-model.number="phaseConfig.final_bonus" min="0" />
          </div>
          
          <div class="setting-group">
            <label>{{ t('settings.phaseConfig.cellsPerRow') }}</label>
            <input type="number" v-model.number="phaseConfig.cells_per_row" min="1" max="5" />
          </div>
          
          <div class="setting-group">
            <label>{{ t('settings.phaseConfig.unlockThreshold') }}</label>
            <input type="number" v-model.number="phaseConfig.unlock_threshold" min="1" max="10" />
          </div>
        </template>

        <div class="setting-group">
          <label>{{ t('settings.roomPassword') }}</label>
          <div class="password-input">
            <input :type="showPassword ? 'text' : 'password'" v-model="roomPassword" :placeholder="t('settings.passwordPlaceholder')" />
            <button @click="showPassword = !showPassword">{{ showPassword ? t('settings.hide') : t('settings.show') }}</button>
          </div>
        </div>
      </div>

      <div class="dialog-footer">
        <button @click="applySettings" :disabled="!canChangeSettings">{{ t('settings.applySettings') }}</button>
        <button v-if="game?.status === 'playing'" @click="handleReset" :disabled="!isOwner">
          {{ t('game.resetBoard') }}
        </button>
        <button v-if="game?.status === 'finished'" @click="handleReset" :disabled="!isOwner" class="restart-btn">
          {{ t('game.restart') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import type { Game, PhaseConfig } from '../types';
import { useGameStore } from '../stores/game';
import { useWebSocket } from '../composables/useWebSocket';
import { useLocaleStore } from '../stores/locale';

const props = defineProps<{
  game: Game | null;
}>();

const emit = defineEmits<{
  (e: 'reset'): void;
}>();

const store = useGameStore();
const { setRule, setPassword } = useWebSocket();
const { t } = useLocaleStore();

const showDialog = ref(false);
const selectedRule = ref('normal');
const showPassword = ref(false);
const roomPassword = ref('');

const defaultPhaseConfig: PhaseConfig = {
  row_scores: [2, 2, 4, 4, 6],
  second_half_rate: 0.5,
  final_bonus: 0,
  final_bonus_type: 'fixed',
  cells_per_row: 3,
  unlock_threshold: 2,
};

const phaseConfig = ref<PhaseConfig>({ ...defaultPhaseConfig });

const isOwner = computed(() => store.isOwner);
const isReferee = computed(() => store.isReferee);
const canChangeSettings = computed(() => isOwner.value && props.game?.status === 'waiting');

watch(() => props.game, (newGame) => {
  if (newGame) {
    selectedRule.value = newGame.rule;
    if (newGame.phase_config) {
      phaseConfig.value = { ...newGame.phase_config };
    }
  }
}, { immediate: true });

function applySettings() {
  setRule(selectedRule.value, phaseConfig.value);
  if (roomPassword.value !== '') {
    setPassword(roomPassword.value);
  }
}

function handleReset() {
  emit('reset');
  showDialog.value = false;
}
</script>

<style scoped>
.settings-btn {
  padding: 8px 16px;
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.settings-btn:hover {
  background: var(--bg-quaternary);
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
  z-index: 1000;
}

.dialog-content {
  background: var(--bg-primary);
  border-radius: 12px;
  width: 400px;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  border-bottom: 1px solid var(--border-light);
}

.dialog-header h3 {
  margin: 0;
  color: var(--text-primary);
  font-size: 18px;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 20px;
  cursor: pointer;
  padding: 0;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  color: var(--text-primary);
}

.dialog-body {
  padding: 20px;
}

.setting-group {
  margin-bottom: 15px;
}

.setting-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
  color: var(--text-secondary);
  font-size: 14px;
}

.setting-group input,
.setting-group select {
  width: 100%;
  padding: 8px;
  border: 1px solid var(--border-light);
  border-radius: 4px;
  box-sizing: border-box;
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.setting-group select {
  background: var(--bg-tertiary);
}

.row-scores {
  display: flex;
  gap: 8px;
}

.row-scores input {
  width: 60px;
  text-align: center;
}

.password-input {
  display: flex;
  gap: 8px;
}

.password-input input {
  flex: 1;
}

.password-input button {
  padding: 8px 12px;
  background: var(--bg-tertiary);
}

.password-input button:hover {
  background: var(--bg-quaternary);
}

.dialog-footer {
  display: flex;
  gap: 10px;
  padding: 15px 20px;
  border-top: 1px solid var(--border-light);
}

.dialog-footer button {
  flex: 1;
  padding: 10px;
}

.restart-btn {
  background: var(--success-color);
}

.restart-btn:hover:not(:disabled) {
  background: var(--success-hover);
}
</style>
