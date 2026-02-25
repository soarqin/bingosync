<template>
  <div class="bingo-board">
    <!-- Owner action buttons (waiting status) -->
    <div v-if="isOwner && game?.status === 'waiting'" class="owner-actions">
      <input 
        type="file" 
        ref="fileInput"
        accept=".txt,.csv"
        @change="handleFileImport"
        style="display: none"
      />
      <button @click="($refs.fileInput as HTMLInputElement).click()" class="action-btn import-btn">
        📁 {{ t('game.importText') }}
      </button>
      <button @click="handleExport" class="action-btn export-btn">
        📤 {{ t('game.exportText') }}
      </button>
      <button @click="startGame" class="action-btn start-btn">
        🎮 {{ t('game.startGame') }}
      </button>
    </div>

    <!-- Referee color picker (available during and after game) -->
    <div v-if="isReferee && (game?.status === 'playing' || game?.status === 'finished')" class="color-picker">
      <span>{{ t('game.selectColor') }}:</span>
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
    
    <div class="board">
      <div v-for="(row, rowIndex) in board.cells" :key="rowIndex" class="row">
        <div
          v-for="(cell, colIndex) in row"
          :key="colIndex"
          class="cell"
          :class="[cell.marked_by, { clickable: canMark || canEditText, locked: isLocked(rowIndex), 'can-edit': canEditText }]"
          @click="handleClick(rowIndex, colIndex)"
        >
          <span 
            class="cell-text" 
            :style="{ fontSize: getCellFontSize(cell.text) }"
          >{{ cell.text }}</span>
          <span v-if="cell.times > 0" class="times">{{ cell.times }}</span>
        </div>
      </div>
    </div>

    <!-- Edit text dialog -->
    <div v-if="showEditDialog" class="dialog-overlay" @click.self="showEditDialog = false">
      <div class="edit-dialog">
        <h4>{{ t('cellEdit.title') }}</h4>
        <textarea 
          ref="editTextarea"
          v-model="editText" 
          :placeholder="t('cellEdit.placeholder')"
          rows="3"
          @keyup.enter.ctrl="saveEditText"
        ></textarea>
        <div class="dialog-actions">
          <button @click="showEditDialog = false">{{ t('common.cancel') }}</button>
          <button @click="saveEditText" class="save-btn">{{ t('common.save') }}</button>
        </div>
      </div>
    </div>

    <!-- Winner display (below board) -->
    <div v-if="game?.status === 'finished' && game.winner" class="winner-display">
      <div class="winner-title">
        <span v-if="game.winner.winner === 'none'" class="tie">{{ t('game.draw') }}!</span>
        <span v-else :class="game.winner.winner">
          {{ game.winner.winner === 'red' ? t('game.redTeam') : t('game.blueTeam') }} {{ t('game.winner') }}!
        </span>
      </div>
      <div class="winner-reason">
        {{ winReasonText }}
      </div>
      <div class="winner-scores">
        <span class="red">{{ t('game.redScore') }}: {{ game.winner.red_score }}</span>
        <span class="blue">{{ t('game.blueScore') }}: {{ game.winner.blue_score }}</span>
      </div>
    </div>

    <div v-if="game" class="game-info">
      <div class="status">
        <span v-if="game.status === 'waiting'">{{ t('game.waiting') }}</span>
        <span v-else-if="game.status === 'playing'">{{ t('game.playing') }}</span>
        <span v-else class="finished">{{ t('game.finished') }}</span>
      </div>
      <div class="counts">
        <span class="red">{{ t('game.redScore') }}: {{ redCount }}</span>
        <span class="blue">{{ t('game.blueScore') }}: {{ blueCount }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import type { Game, Board, PlayerColor, WinReason } from '../types';
import { useGameStore } from '../stores/game';
import { useWebSocket } from '../composables/useWebSocket';
import { useLocaleStore } from '../stores/locale';

// Cell size constants
const CELL_SIZE = 80;
const CELL_PADDING = 6; // Consistent with CSS padding
const LINE_HEIGHT = 1.3;
const MIN_FONT_SIZE = 7;
const MAX_FONT_SIZE = 18;
const SAFETY_FACTOR = 0.92; // Safety factor to prevent overflow due to estimation errors

const props = defineProps<{
  board: Board;
  game: Game | null;
}>();

const emit = defineEmits<{
  (e: 'mark', row: number, col: number, color: PlayerColor): void;
}>();

const store = useGameStore();
const { setCellText, setAllCellTexts, startGame } = useWebSocket();
const { t } = useLocaleStore();

// Calculate estimated width of a string (in font size units)
// Use conservative estimation to ensure width is not underestimated
function estimateTextWidth(text: string): number {
  let width = 0;
  for (const char of text) {
    const code = char.charCodeAt(0);
    if (code >= 0x4e00 && code <= 0x9fff) {
      // Chinese characters
      width += 1.0;
    } else if (code >= 0x3040 && code <= 0x30ff) {
      // Japanese characters
      width += 1.0;
    } else if (code >= 0xac00 && code <= 0xd7af) {
      // Korean characters
      width += 1.0;
    } else if (code >= 0x00 && code <= 0x7f) {
      // ASCII characters (letters, numbers, symbols)
      if (char === '\n') {
        width += 0; // Newline doesn't count for width
      } else if (char >= '0' && char <= '9') {
        width += 0.6;
      } else if ((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
        width += 0.65;
      } else {
        width += 0.5;
      }
    } else {
      // Other characters (emoji, etc.)
      width += 1.2;
    }
  }
  return width;
}

// Calculate the number of lines after text wrapping based on font size
function calculateWrappedLines(text: string, fontSize: number, availableWidth: number): number {
  const explicitLines = text.split('\n');
  let totalLines = 0;
  
  // Maximum line width in font size units
  const maxLineWidth = availableWidth / fontSize;
  
  for (const line of explicitLines) {
    if (line.length === 0) {
      totalLines += 1;
      continue;
    }
    
    const lineWidth = estimateTextWidth(line);
    // Calculate how many lines this segment needs (rounded up)
    const wrappedLines = Math.ceil(lineWidth / maxLineWidth);
    totalLines += Math.max(1, wrappedLines);
  }
  
  return totalLines;
}

// Calculate the optimal font size for a cell (using binary search)
function getCellFontSize(text: string): string {
  if (!text) return `${MAX_FONT_SIZE}px`;
  
  // Calculate available width and height (considering padding)
  const availableWidth = (CELL_SIZE - CELL_PADDING * 2) * SAFETY_FACTOR;
  const availableHeight = (CELL_SIZE - CELL_PADDING * 2) * SAFETY_FACTOR;
  
  // Binary search to find the maximum feasible font size
  let low = MIN_FONT_SIZE;
  let high = MAX_FONT_SIZE;
  let bestSize = MIN_FONT_SIZE;
  
  // Precision to 0.5px
  while (high - low > 0.5) {
    const mid = (low + high) / 2;
    
    // Calculate total lines at this font size (considering text wrapping)
    const totalLines = calculateWrappedLines(text, mid, availableWidth);
    
    // Calculate total height
    const totalHeight = totalLines * mid * LINE_HEIGHT;
    
    if (totalHeight <= availableHeight) {
      // Height is sufficient, try larger font
      bestSize = mid;
      low = mid;
    } else {
      // Height is not enough, need smaller font
      high = mid;
    }
  }
  
  return `${Math.round(bestSize)}px`;
}

// Referee selected color
const selectedColor = ref<PlayerColor>('red');

// Text editing related
const showEditDialog = ref(false);
const editText = ref('');
const editRow = ref(0);
const editCol = ref(0);
const editTextarea = ref<HTMLTextAreaElement | null>(null);

// Watch dialog open, auto focus
watch(showEditDialog, async (newValue) => {
  if (newValue) {
    await nextTick();
    editTextarea.value?.focus();
  }
});

const isOwner = computed(() => store.isOwner);
const isReferee = computed(() => store.isReferee);

// Whether text can be edited (owner and game is waiting)
const canEditText = computed(() => {
  return isOwner.value && props.game?.status === 'waiting';
});

const canMark = computed(() => {
  if (!props.game) return false;
  
  // Cannot mark while game is waiting (but owner can edit text)
  if (props.game.status === 'waiting') return false;
  
  // During game: referee and players can mark
  if (props.game.status === 'playing') {
    if (store.isReferee) return true;
    if (store.isPlayer) return true;
    return false;
  }
  
  // After game finished: only referee can mark
  if (props.game.status === 'finished') {
    return store.isReferee;
  }
  
  return false;
});

const redCount = computed(() => {
  let count = 0;
  for (const row of props.board.cells) {
    for (const cell of row) {
      if (cell.marked_by === 'red') count++;
    }
  }
  return count;
});

const blueCount = computed(() => {
  let count = 0;
  for (const row of props.board.cells) {
    for (const cell of row) {
      if (cell.marked_by === 'blue') count++;
    }
  }
  return count;
});

const winReasonText = computed(() => {
  if (!props.game?.winner) return '';
  const reason: WinReason = props.game.winner.reason;
  switch (reason) {
    case 'bingo':
      return t('winReason.bingo');
    case 'full_board':
      return t('winReason.fullBoard');
    case 'blackout':
      return t('winReason.blackout');
    default:
      return '';
  }
});

function isLocked(row: number): boolean {
  if (!props.game || props.game.rule !== 'phase') return false;
  return row > (props.game.current_row ?? 0);
}

function handleClick(row: number, col: number) {
  // Owner can edit text while game is waiting
  if (canEditText.value) {
    editRow.value = row;
    editCol.value = col;
    editText.value = props.board.cells[row][col].text || '';
    showEditDialog.value = true;
    return;
  }

  if (!canMark.value) return;
  if (isLocked(row)) return;
  
  const cell = props.board.cells[row][col];
  
  // Referee mode
  if (store.isReferee) {
    // Use referee selected color
    emit('mark', row, col, selectedColor.value);
    return;
  }
  
  // Player mode: can only mark own color
  const playerColor = store.currentUser?.player_color;
  if (playerColor && playerColor !== 'none') {
    // Normal rule: cannot mark already marked cell (unless canceling own color)
    if (props.game?.rule === 'normal' && cell.marked_by !== 'none') {
      // If marked by self, can cancel
      if (cell.marked_by === playerColor) {
        emit('mark', row, col, 'none');
      }
      return;
    }
    emit('mark', row, col, playerColor);
  }
}

function saveEditText() {
  setCellText(editRow.value, editCol.value, editText.value);
  showEditDialog.value = false;
}

async function handleFileImport(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0];
  if (!file) return;
  
  try {
    const text = await file.text();
    let texts: string[] = [];
    
    if (file.name.toLowerCase().endsWith('.txt')) {
      // TXT: one per line, 25 lines total (supports \\n for cell line breaks)
      const lines = text.split('\n').map(line => {
        // Convert \\n to actual newline character
        return line.trim().replace(/\\n/g, '\n');
      }).filter(line => line.length > 0);
      texts = lines.slice(0, 25);
    } else if (file.name.toLowerCase().endsWith('.csv')) {
      // CSV: 5 per line, 5 lines total (supports quoted text with line breaks)
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
    
    setAllCellTexts(texts);
  } catch (e) {
    console.error('Failed to import file:', e);
    store.setError(t('settings.importFailed'));
  }
  
  // Reset file input
  (event.target as HTMLInputElement).value = '';
}

// Parse CSV line, supports quoted text
function parseCSVLine(line: string): string[] {
  const result: string[] = [];
  let current = '';
  let inQuotes = false;
  
  for (let i = 0; i < line.length; i++) {
    const char = line[i];
    
    if (char === '"') {
      if (inQuotes && line[i + 1] === '"') {
        // Escaped quote
        current += '"';
        i++;
      } else {
        // Toggle quote state
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
  // Collect all cell texts
  const texts: string[] = [];
  for (const row of props.board.cells) {
    for (const cell of row) {
      texts.push(cell.text || '');
    }
  }
  
  // Generate CSV content (using quotes to support line breaks)
  const csvLines: string[] = [];
  for (let i = 0; i < 5; i++) {
    const rowTexts = texts.slice(i * 5, (i + 1) * 5);
    const csvRow = rowTexts.map(text => {
      // If text contains comma, quote or newline, wrap with quotes
      if (text.includes(',') || text.includes('"') || text.includes('\n')) {
        // Escape quotes
        return '"' + text.replace(/"/g, '""') + '"';
      }
      return text;
    }).join(',');
    csvLines.push(csvRow);
  }
  
  const csvContent = csvLines.join('\n');
  
  // Create and download file
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
</script>

<style scoped>
.bingo-board {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

/* Owner action buttons */
.owner-actions {
  display: flex;
  gap: 15px;
  padding: 15px;
  background: var(--bg-tertiary);
  border-radius: 8px;
}

.action-btn {
  padding: 12px 24px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.2s;
}

.import-btn {
  background: var(--info-color);
  color: var(--text-on-accent);
}

.import-btn:hover {
  background: var(--info-hover);
}

.start-btn {
  background: var(--success-color);
  color: var(--text-on-accent);
}

.start-btn:hover {
  background: var(--success-hover);
  transform: scale(1.05);
}

.export-btn {
  background: #9b59b6;
  color: var(--text-on-accent);
}

.export-btn:hover {
  background: #8e44ad;
}

.winner-display {
  padding: 20px;
  background: linear-gradient(135deg, var(--bg-tertiary) 0%, var(--bg-primary) 100%);
  border-radius: 12px;
  text-align: center;
  min-width: 300px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

.winner-title {
  font-size: 28px;
  font-weight: bold;
  margin-bottom: 10px;
}

.winner-title .red {
  color: var(--red-color);
  text-shadow: 0 0 10px rgba(231, 76, 60, 0.5);
}

.winner-title .blue {
  color: var(--blue-color);
  text-shadow: 0 0 10px rgba(52, 152, 219, 0.5);
}

.winner-title .tie {
  color: var(--warning-color);
  text-shadow: 0 0 10px rgba(243, 156, 18, 0.5);
}

.winner-reason {
  font-size: 16px;
  color: var(--text-muted);
  margin-bottom: 15px;
}

.winner-scores {
  display: flex;
  justify-content: center;
  gap: 30px;
  font-size: 18px;
}

/* 编辑文字对话框 */
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

.edit-dialog {
  background: var(--bg-primary);
  border-radius: 12px;
  padding: 20px;
  width: 300px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
}

.edit-dialog h4 {
  margin: 0 0 15px 0;
  color: var(--text-primary);
}

.edit-dialog textarea {
  width: 100%;
  height: 80px;
  padding: 10px;
  border: 1px solid var(--border-light);
  border-radius: 4px;
  background: var(--bg-tertiary);
  color: var(--text-primary);
  resize: none;
  font-size: 14px;
  box-sizing: border-box;
}

.edit-dialog textarea:focus {
  outline: none;
  border-color: var(--accent-color);
}

.dialog-actions {
  display: flex;
  gap: 10px;
  margin-top: 15px;
}

.dialog-actions button {
  flex: 1;
  padding: 10px;
}

.dialog-actions .save-btn {
  background: var(--success-color);
}

.dialog-actions .save-btn:hover {
  background: var(--success-hover);
}

.color-picker {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  background: var(--bg-tertiary);
  border-radius: 8px;
}

.color-picker span {
  color: var(--text-primary);
  font-size: 14px;
}

.color-btn {
  padding: 8px 16px;
  border: 2px solid transparent;
  border-radius: 4px;
  cursor: pointer;
  font-weight: bold;
  transition: all 0.2s;
}

.color-btn.red {
  background: var(--red-color);
  color: var(--text-on-accent);
}

.color-btn.blue {
  background: var(--blue-color);
  color: var(--text-on-accent);
}

.color-btn.clear {
  background: var(--clear-btn-bg);
  color: var(--text-on-accent);
}

.color-btn.active {
  border-color: var(--text-primary);
  transform: scale(1.1);
}

.board {
  display: flex;
  flex-direction: column;
  gap: 4px;
  background: var(--bg-quaternary);
  padding: 8px;
  border-radius: 8px;
}

.row {
  display: flex;
  gap: 4px;
}

.cell {
  width: 80px;
  height: 80px;
  min-height: 80px;
  max-height: 80px;
  background: var(--cell-bg);
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  transition: all 0.2s;
  padding: 6px;
  box-sizing: border-box;
}

.cell-text {
  color: var(--cell-text);
  text-align: center;
  word-break: break-all;
  line-height: 1.3;
  overflow: hidden;
  white-space: pre-wrap;
  max-width: 100%;
  max-height: 100%;
}

.cell.red .cell-text,
.cell.blue .cell-text {
  color: var(--text-on-accent);
}

.cell.clickable {
  cursor: pointer;
}

.cell.clickable:hover {
  transform: scale(1.05);
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.3);
}

.cell.can-edit {
  cursor: pointer;
  border: 2px dashed var(--border-light);
}

.cell.can-edit:hover {
  border-color: var(--accent-color);
  transform: scale(1.02);
}

.cell.locked {
  opacity: 0.5;
}

.cell.red {
  background: var(--red-color);
}

.cell.blue {
  background: var(--blue-color);
}

.cell.none {
  background: var(--cell-empty-bg);
}

.times {
  position: absolute;
  bottom: 2px;
  right: 4px;
  font-size: 10px;
  color: var(--cell-times-color);
}

.cell.red .times,
.cell.blue .times {
  color: var(--text-on-accent);
}

.game-info {
  text-align: center;
}

.status {
  font-size: 18px;
  margin-bottom: 8px;
  color: var(--text-primary);
}

.status .finished {
  color: var(--warning-color);
}

.counts {
  display: flex;
  gap: 20px;
  font-size: 16px;
  color: var(--text-primary);
}

.red {
  color: var(--red-color);
}

.blue {
  color: var(--blue-color);
}
</style>
