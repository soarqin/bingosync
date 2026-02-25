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
              <input v-for="i in 5" :key="'a'+i" type="number" v-model.number="phaseConfig.row_scores[i-1]" min="0" />
            </div>
          </div>
          
          <div class="setting-group">
            <label>{{ t('settings.phaseConfig.secondHalfScores') }}</label>
            <div class="row-scores">
              <input v-for="i in 5" :key="'b'+i" type="number" v-model.number="phaseConfig.second_half_scores[i-1]" min="0" />
            </div>
          </div>
          
          <div class="setting-group">
            <label>{{ t('settings.phaseConfig.cellsPerRow') }}</label>
            <input type="number" v-model.number="phaseConfig.cells_per_row" min="1" max="5" />
          </div>
          
          <div class="setting-group">
            <label>{{ t('settings.phaseConfig.unlockThreshold') }}</label>
            <input type="number" v-model.number="phaseConfig.unlock_threshold" min="1" max="10" />
          </div>
          
          <div class="setting-group">
            <label>{{ t('settings.phaseConfig.bingoBonus') }}</label>
            <input type="number" v-model.number="phaseConfig.bingo_bonus" min="0" />
          </div>
          
          <div class="setting-group">
            <label>{{ t('settings.phaseConfig.finalBonus') }}</label>
            <input type="number" v-model.number="phaseConfig.final_bonus" min="0" />
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
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import type { Game, PhaseConfig } from '../types';
import { useGameStore } from '../stores/game';
import { useWebSocket } from '../composables/useWebSocket';
import { useLocaleStore } from '../stores/locale';

const props = defineProps<{
  game: Game | null;
}>();

const store = useGameStore();
const { setRule, setPassword } = useWebSocket();
const { t } = useLocaleStore();

const STORAGE_KEY_RULE = 'bingosync-settings-rule';
const STORAGE_KEY_PHASE_CONFIG = 'bingosync-settings-phase-config';

const showDialog = ref(false);
const selectedRule = ref('normal');
const showPassword = ref(false);
const roomPassword = ref('');
const lastAppliedRoomId = ref<string | null>(null);

const defaultPhaseConfig: PhaseConfig = {
  row_scores: [2, 2, 4, 4, 6],
  second_half_scores: [1, 1, 2, 2, 3],
  cells_per_row: 3,
  unlock_threshold: 2,
  bingo_bonus: 3,
  final_bonus: 3,
};

const phaseConfig = ref<PhaseConfig>({ ...defaultPhaseConfig });

const isOwner = computed(() => store.isOwner);
const canChangeSettings = computed(() => isOwner.value && props.game?.status === 'waiting');

// Load settings from localStorage on mount
onMounted(() => {
  loadSettingsFromStorage();
});

// Watch for game changes to sync with current game state
watch(() => props.game, (newGame) => {
  if (newGame) {
    selectedRule.value = newGame.rule;
    if (newGame.phase_config) {
      phaseConfig.value = { ...newGame.phase_config };
    }
  }
}, { immediate: true });

// Auto-apply saved settings when entering a waiting room as owner
watch([() => store.currentRoom, () => props.game?.status, () => store.isOwner], 
  ([room, status, isOwner]) => {
    if (!room || status !== 'waiting' || !isOwner) return;
    
    // Only apply once per room
    if (lastAppliedRoomId.value === room.id) return;
    lastAppliedRoomId.value = room.id;
    
    // Load and apply saved settings
    const savedRule = localStorage.getItem(STORAGE_KEY_RULE);
    const savedPhaseConfig = localStorage.getItem(STORAGE_KEY_PHASE_CONFIG);
    
    if (savedRule || savedPhaseConfig) {
      loadSettingsFromStorage();
      // Apply settings after a short delay to ensure room is ready
      setTimeout(() => {
        setRule(selectedRule.value, phaseConfig.value);
      }, 100);
    }
  },
  { immediate: true }
);

function loadSettingsFromStorage() {
  try {
    // Load rule
    const savedRule = localStorage.getItem(STORAGE_KEY_RULE);
    if (savedRule && ['normal', 'blackout', 'phase'].includes(savedRule)) {
      selectedRule.value = savedRule;
    }

    // Load phase config
    const savedPhaseConfig = localStorage.getItem(STORAGE_KEY_PHASE_CONFIG);
    if (savedPhaseConfig) {
      const parsed = JSON.parse(savedPhaseConfig);
      if (parsed && typeof parsed === 'object') {
        phaseConfig.value = { ...defaultPhaseConfig, ...parsed };
      }
    }
  } catch (e) {
    console.error('Failed to load settings from localStorage:', e);
  }
}

function saveSettingsToStorage() {
  try {
    localStorage.setItem(STORAGE_KEY_RULE, selectedRule.value);
    localStorage.setItem(STORAGE_KEY_PHASE_CONFIG, JSON.stringify(phaseConfig.value));
  } catch (e) {
    console.error('Failed to save settings to localStorage:', e);
  }
}

function applySettings() {
  // Apply settings to game
  setRule(selectedRule.value, phaseConfig.value);
  if (roomPassword.value !== '') {
    setPassword(roomPassword.value);
  }

  // Save settings to localStorage (except password)
  saveSettingsToStorage();

  // Close the dialog
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
