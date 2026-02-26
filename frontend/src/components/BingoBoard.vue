<template>
  <div class="bingo-board">
    <!-- Bingo notification (floating) -->
    <div v-if="game?.rule === 'phase' && game?.status === 'playing' && game.bingo_achiever && game.bingo_achiever !== 'none'" class="bingo-notification">
      🎉 {{ game.bingo_achiever === 'red' ? t('game.redTeam') : t('game.blueTeam') }} {{ t('phase.bingoAchieved') }}! 🎉
    </div>

    <div class="board" ref="boardRef">
      <div v-for="(row, rowIndex) in board.cells" :key="rowIndex" class="row">
        <div
          v-for="(cell, colIndex) in row"
          :key="colIndex"
          class="cell"
          :class="getCellClass(cell, rowIndex)"
          @click="handleClick(rowIndex, colIndex)"
          @contextmenu="handleRightClick($event, rowIndex, colIndex)"
        >
          <span 
            class="cell-text" 
            :style="{ fontSize: getCellFontSize(cell.text) }"
          >{{ cell.text }}</span>
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

    <!-- Game info section below board (always visible) -->
    <div v-if="game" class="game-info-section">
      <!-- Scores (top priority) -->
      <div class="scores-row">
        <span class="red">{{ t('game.redScore') }}: {{ redCount }}</span>
        <span class="blue">{{ t('game.blueScore') }}: {{ blueCount }}</span>
      </div>
      
      <!-- Game status -->
      <div class="status-row">
        <span v-if="game.status === 'waiting'">{{ t('game.waiting') }}</span>
        <span v-else-if="game.status === 'playing'">{{ t('game.playing') }}</span>
        <span v-else class="finished">{{ t('game.finished') }}</span>
      </div>
    </div>

    <!-- Winner display -->
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, inject, type Ref, onMounted, onUnmounted } from 'vue';
import type { Game, Board, PlayerColor, WinReason } from '../types';
import { useGameStore } from '../stores/game';
import { useWebSocket } from '../composables/useWebSocket';
import { useLocaleStore } from '../stores/locale';

// Cell size constants
const BASE_CELL_SIZE = 80; // Base cell size in pixels
const CELL_PADDING = 6; // Consistent with CSS padding
const LINE_HEIGHT = 1.3;
const MIN_FONT_SIZE = 7;
const MAX_FONT_SIZE = 18;
const SAFETY_FACTOR = 0.92; // Safety factor to prevent overflow due to estimation errors

// Responsive cell size
const cellSize = ref(BASE_CELL_SIZE);
const boardRef = ref<HTMLElement | null>(null);

const props = defineProps<{
  board: Board;
  game: Game | null;
}>();

const emit = defineEmits<{
  (e: 'mark', row: number, col: number, color: PlayerColor): void;
  (e: 'settle', color: PlayerColor): void;
}>();

const store = useGameStore();
const { setCellText, setAllCellTexts, startGame, settle, clearCellMark } = useWebSocket();
const { t } = useLocaleStore();

// Inject selectedColor from parent (App.vue)
const selectedColor = inject<Ref<PlayerColor>>('selectedColor');
if (!selectedColor) {
  console.error('selectedColor not provided from parent');
}

// Update cell size based on actual rendered cell
function updateCellSize() {
  if (!boardRef.value) return;
  
  // Get the first cell's actual rendered size
  const firstCell = boardRef.value.querySelector('.cell');
  if (!firstCell) return;
  
  const rect = firstCell.getBoundingClientRect();
  const size = rect.width;
  
  // Skip if size is too small or not ready
  if (size < 30) return;
  
  cellSize.value = Math.round(size);
}

// Resize observer
let resizeObserver: ResizeObserver | null = null;

onMounted(() => {
  // Initial calculation
  const initUpdate = () => {
    nextTick(() => {
      updateCellSize();
    });
  };
  
  // Try immediately and after delays
  initUpdate();
  setTimeout(initUpdate, 100);
  setTimeout(initUpdate, 300);
  
  // Use ResizeObserver to update cell size when board resizes
  if (boardRef.value) {
    resizeObserver = new ResizeObserver(() => {
      updateCellSize();
    });
    resizeObserver.observe(boardRef.value);
  }
  
  // Fallback to window resize
  window.addEventListener('resize', updateCellSize);
});

onUnmounted(() => {
  if (resizeObserver) {
    resizeObserver.disconnect();
  }
  window.removeEventListener('resize', updateCellSize);
});

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
  const availableWidth = (cellSize.value - CELL_PADDING * 2) * SAFETY_FACTOR;
  const availableHeight = (cellSize.value - CELL_PADDING * 2) * SAFETY_FACTOR;
  
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

const redCount = computed(() => calculateScores().red);

const blueCount = computed(() => calculateScores().blue);

// Calculate scores for both players
function calculateScores(): { red: number; blue: number } {
  // For phase rule, calculate actual score
  if (props.game?.rule === 'phase' && props.game.phase_config) {
    return {
      red: calculatePhaseScore('red'),
      blue: calculatePhaseScore('blue')
    };
  }
  // For other rules, count marked cells
  let red = 0, blue = 0;
  for (const row of props.board.cells) {
    for (const cell of row) {
      if (cell.marked_by === 'red') red++;
      if (cell.marked_by === 'blue') blue++;
    }
  }
  return { red, blue };
}

// Calculate phase rule score for a player
function calculatePhaseScore(color: 'red' | 'blue'): number {
  if (!props.game?.phase_config) return 0;
  
  const config = props.game.phase_config;
  let score = 0;
  
  for (let row = 0; row < 5; row++) {
    for (let col = 0; col < 5; col++) {
      const cell = props.board.cells[row][col];
      
      // First marker gets full row score
      if (cell.marked_by === color) {
        score += config.row_scores[row];
      }
      // Second marker gets reduced score
      if (cell.second_mark === color) {
        score += config.second_half_scores[row];
      }
    }
  }
  
  // Add bingo bonus
  if (props.game.bingo_achiever === color) {
    score += config.bingo_bonus;
  }
  
  // Add first settler bonus
  if (props.game.first_settler === color) {
    score += config.final_bonus;
  }
  
  return score;
}

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
    case 'phase':
      return t('winReason.phase');
    default:
      return '';
  }
});

// Phase rule: check if current player can settle
const canSettle = computed(() => {
  if (!props.game || props.game.rule !== 'phase') return false;
  if (props.game.status !== 'playing') return false;
  return store.isPlayer || store.isReferee;
});

// Check if current player has already settled
const isCurrentPlayerSettled = computed(() => {
  if (!props.game) return false;
  const color = store.currentUser?.player_color;
  if (color === 'red') return props.game.red_settled ?? false;
  if (color === 'blue') return props.game.blue_settled ?? false;
  return false;
});

// Check if player meets settlement conditions (>= 2 cells in row 5)
const canSettleNow = computed(() => {
  if (!props.game) return false;
  const color = store.currentUser?.player_color;
  if (color === 'red') {
    return (props.game.red_row_marks?.[4] ?? 0) >= 2;
  }
  if (color === 'blue') {
    return (props.game.blue_row_marks?.[4] ?? 0) >= 2;
  }
  return false;
});

// For referee: check if each player can settle
const canRedSettle = computed(() => {
  if (!props.game) return false;
  return (props.game.red_row_marks?.[4] ?? 0) >= 2;
});

const canBlueSettle = computed(() => {
  if (!props.game) return false;
  return (props.game.blue_row_marks?.[4] ?? 0) >= 2;
});

function handleSettle() {
  const color = store.currentUser?.player_color;
  if (color && color !== 'none') {
    emit('settle', color);
  }
}

function handleRefereeSettle(color: 'red' | 'blue') {
  emit('settle', color);
}

// Expose settle-related properties and methods to parent component
defineExpose({
  canSettle,
  isCurrentPlayerSettled,
  canSettleNow,
  canRedSettle,
  canBlueSettle,
  handleSettle,
  handleRefereeSettle
});

function isLocked(row: number): boolean {
  if (!props.game || props.game.rule !== 'phase') return false;
  
  // Get player's unlocked row
  const playerColor = store.currentUser?.player_color;
  let unlockedRow = 0;
  
  if (store.isReferee) {
    // Referee can mark any unlocked row (use max of both players)
    unlockedRow = Math.max(
      props.game.red_unlocked_row ?? 0,
      props.game.blue_unlocked_row ?? 0
    );
  } else if (playerColor === 'red') {
    unlockedRow = props.game.red_unlocked_row ?? 0;
  } else if (playerColor === 'blue') {
    unlockedRow = props.game.blue_unlocked_row ?? 0;
  } else {
    // Spectator: use max of both players for display purposes
    unlockedRow = Math.max(
      props.game.red_unlocked_row ?? 0,
      props.game.blue_unlocked_row ?? 0
    );
  }
  
  return row > unlockedRow;
}

function getCellClass(cell: { marked_by: string; second_mark?: string }, row: number): Record<string, boolean> {
  const classes: Record<string, boolean> = {
    clickable: canMark.value || canEditText.value,
    locked: isLocked(row),
    'can-edit': canEditText.value,
  };
  
  // For blackout and phase rules with second mark, use diagonal pattern
  if ((props.game?.rule === 'phase' || props.game?.rule === 'blackout') && 
      cell.second_mark && cell.second_mark !== 'none') {
    classes['both-marks'] = true;
    classes[cell.marked_by] = true;
    classes[`second-${cell.second_mark}`] = true;
  } else {
    classes[cell.marked_by] = true;
  }
  
  return classes;
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
    emit('mark', row, col, selectedColor?.value || 'none');
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
    
    // Blackout and phase rules: can mark if not already marked by this player
    if (props.game?.rule === 'blackout' || props.game?.rule === 'phase') {
      // If cell already has this player's mark, do nothing
      if (cell.marked_by === playerColor || cell.second_mark === playerColor) {
        return;
      }
    }
    
    emit('mark', row, col, playerColor);
  }
}

// Right-click handler to clear specific color
function handleRightClick(event: MouseEvent, row: number, col: number) {
  event.preventDefault();
  
  if (!canMark.value) return;
  
  const cell = props.board.cells[row][col];
  
  // Only for blackout and phase rules
  if (props.game?.rule !== 'blackout' && props.game?.rule !== 'phase') {
    return;
  }
  
  // Referee can clear any color
  if (store.isReferee) {
    // Clear the selected color
    if (selectedColor?.value && selectedColor.value !== 'none') {
      clearCellMark(row, col, selectedColor.value);
    }
    return;
  }
  
  // Player can only clear their own color
  const playerColor = store.currentUser?.player_color;
  if (playerColor && playerColor !== 'none') {
    // Check if this player has a mark on the cell
    if (cell.marked_by === playerColor || cell.second_mark === playerColor) {
      clearCellMark(row, col, playerColor);
    }
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
  position: relative;
}

.bingo-notification {
  position: absolute;
  top: -60px;
  left: 50%;
  transform: translateX(-50%);
  background: linear-gradient(135deg, #f39c12, #e67e22);
  color: white;
  padding: 12px 24px;
  border-radius: 25px;
  font-size: 18px;
  font-weight: bold;
  box-shadow: 0 4px 20px rgba(243, 156, 18, 0.5);
  animation: pulse 2s infinite;
  z-index: 100;
  white-space: nowrap;
}

@keyframes pulse {
  0%, 100% {
    transform: translateX(-50%) scale(1);
  }
  50% {
    transform: translateX(-50%) scale(1.05);
  }
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

.bingo-board {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 15px;
  width: 100%;
  height: 100%;
  min-height: 0;
  box-sizing: border-box;
}

.board {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  grid-template-rows: repeat(5, 1fr);
  gap: 4px;
  background: var(--bg-quaternary);
  padding: 8px;
  border-radius: 8px;
  /* Board size is limited by container width and available height */
  /* Use CSS custom property for max-height calculation */
  width: 100%;
  max-width: min(600px, calc(100vh - 280px));
  aspect-ratio: 1;
  flex-shrink: 1;
  flex-grow: 0;
  margin: 0 auto;
  box-sizing: border-box;
}

.row {
  display: contents;
}

.cell {
  background: var(--cell-bg);
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  padding: 6px;
  box-sizing: border-box;
  transition: transform 0.15s ease, background-color 0.15s ease;
}

.cell.clickable {
  cursor: pointer;
}

.cell.clickable:hover {
  transform: scale(1.05);
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.3);
  z-index: 1;
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

/* Cell with both marks - show second mark as bottom quarter solid fill */
.cell.both-marks {
  position: relative;
}

/* Bottom quarter solid fill for second mark */
.cell.both-marks.second-red::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 25%;
  background: rgba(231, 76, 60, 0.9);
  border-radius: 0 0 4px 4px;
  pointer-events: none;
}

.cell.both-marks.second-blue::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 25%;
  background: rgba(52, 152, 219, 0.9);
  border-radius: 0 0 4px 4px;
  pointer-events: none;
}

/* Ensure text is still visible on both-marks cells */
.cell.both-marks .cell-text {
  position: relative;
  z-index: 1;
}

/* Game info section below board */
.game-info-section {
  text-align: center;
  margin-top: 15px;
  padding: 10px;
  background: var(--bg-tertiary);
  border-radius: 8px;
  width: 100%;
  max-width: min(600px, calc(100vh - 280px));
  box-sizing: border-box;
}

.scores-row {
  display: flex;
  justify-content: center;
  gap: 30px;
  font-size: 18px;
  font-weight: bold;
  margin-bottom: 8px;
}

.scores-row .red {
  color: var(--red-color);
}

.scores-row .blue {
  color: var(--blue-color);
}

.status-row {
  font-size: 14px;
  color: var(--text-secondary);
}

.status-row .finished {
  color: var(--warning-color);
  font-weight: bold;
}

.red {
  color: var(--red-color);
}

.blue {
  color: var(--blue-color);
}
</style>
