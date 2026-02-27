// Types for BingoSync

// Protocol version - must match server's ProtocolVersion
export const PROTOCOL_VERSION = 1;

export type GameRule = 'normal' | 'blackout' | 'phase';
export type GameStatus = 'waiting' | 'playing' | 'finished';
export type PlayerColor = 'none' | 'red' | 'blue';
export type UserRole = 'spectator' | 'player' | 'referee';
export type WinReason = 'bingo' | 'full_board' | 'blackout' | 'phase';

export interface Cell {
  marked_by: PlayerColor;
  second_mark?: PlayerColor;
  times: number;
  text: string;
}

export interface Board {
  cells: Cell[][];
}

export interface PhaseConfig {
  row_scores: number[];
  second_half_scores: number[];
  cells_per_row: number;
  unlock_threshold: number;
  bingo_bonus: number;
  final_bonus: number;
}

export interface Winner {
  winner: PlayerColor;
  reason: WinReason;
  red_score: number;
  blue_score: number;
}

export interface Game {
  board: Board;
  rule: GameRule;
  phase_config?: PhaseConfig;
  status: GameStatus;
  winner?: Winner;
  red_row_marks?: number[];
  blue_row_marks?: number[];
  red_unlocked_row?: number;
  blue_unlocked_row?: number;
  bingo_achiever?: PlayerColor;
  bingo_line?: number;
  red_settled?: boolean;
  blue_settled?: boolean;
  first_settler?: PlayerColor;
}

export interface User {
  id: string;
  name: string;
  role: UserRole;
  player_color: PlayerColor;
}

export interface Room {
  id: string;
  name: string;
  owner_id: string;
  has_password: boolean;
}

export interface RoomInfo {
  id: string;
  name: string;
  has_password: boolean;
  player_count: number;
  owner_name: string;
}

export interface StateUpdate {
  room: Room;
  game: Game;
  users: User[];
  current_user: string;
}

// Message types
export type MessageType =
  | 'set_name'
  | 'create_room'
  | 'join_room'
  | 'leave_room'
  | 'set_role'
  | 'list_rooms'
  | 'set_password'
  | 'mark_cell'
  | 'unmark_cell'
  | 'clear_cell_mark'
  | 'set_rule'
  | 'start_game'
  | 'reset_game'
  | 'set_cell_text'
  | 'settle'
  | 'state_update'
  | 'room_list'
  | 'error'
  | 'joined'
  | 'left'
  | 'connected'
  | 'name_set';

export interface Message {
  type: MessageType;
  room_id?: string;
  user_id?: string;
  payload?: unknown;
}

export interface ErrorPayload {
  code: number;
  message: string;
}
