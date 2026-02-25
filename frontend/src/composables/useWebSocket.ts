import type { Message, StateUpdate, RoomInfo, ErrorPayload } from '../types';
import { useGameStore } from '../stores/game';

export function useWebSocket() {
  const store = useGameStore();

  function connect(url: string): Promise<void> {
    return new Promise((resolve, reject) => {
      if (store.ws?.readyState === WebSocket.OPEN) {
        resolve();
        return;
      }

      store.connecting = true;
      const socket = new WebSocket(url);

      socket.onopen = () => {
        store.ws = socket;
        store.connecting = false;
        store.setConnected(true);
        resolve();
      };

      socket.onclose = () => {
        store.ws = null;
        store.connecting = false;
        store.reset();
      };

      socket.onerror = (err) => {
        store.connecting = false;
        reject(err);
      };

      socket.onmessage = (event) => {
        handleMessage(event.data);
      };
    });
  }

  function disconnect() {
    if (store.ws) {
      store.ws.close();
      store.ws = null;
    }
  }

  function handleMessage(data: string) {
    try {
      const msg: Message = JSON.parse(data);
      
      switch (msg.type) {
        case 'connected':
          if (msg.payload) {
            const payload = msg.payload as { user_id: string; user_name: string };
            store.setUserInfo(payload.user_id, payload.user_name);
          }
          break;
          
        case 'state_update':
          if (msg.payload) {
            store.setStateUpdate(msg.payload as StateUpdate);
          }
          break;
          
        case 'room_list':
          if (msg.payload) {
            const payload = msg.payload as { rooms: RoomInfo[] };
            store.setRoomList(payload.rooms);
          }
          break;
          
        case 'joined':
          if (msg.payload) {
            store.setStateUpdate(msg.payload as StateUpdate);
          }
          break;
          
        case 'left':
          store.leaveRoom();
          break;
          
        case 'error':
          if (msg.payload) {
            const payload = msg.payload as ErrorPayload;
            store.setError(payload.message);
          }
          break;
        
        case 'name_set':
          if (msg.payload) {
            const payload = msg.payload as { user_name: string };
            store.setUserInfo(store.userId, payload.user_name);
          }
          break;
      }
    } catch (e) {
      console.error('Failed to parse message:', e);
    }
  }

  function send(type: Message['type'], payload?: unknown) {
    if (store.ws?.readyState === WebSocket.OPEN) {
      const msg: Message = { type, payload };
      store.ws.send(JSON.stringify(msg));
    } else {
      console.warn('WebSocket not connected, cannot send:', type);
    }
  }

  // Room actions
  function createRoom(name: string, password?: string) {
    send('create_room', { name, password, user_name: store.userName });
  }

  function joinRoom(roomId: string, password?: string) {
    send('join_room', { room_id: roomId, password, user_name: store.userName });
  }

  function leaveRoom() {
    send('leave_room');
  }

  function listRooms() {
    send('list_rooms');
  }

  function setRole(targetUserId: string, role: string, playerColor?: string) {
    send('set_role', { target_user_id: targetUserId, role, player_color: playerColor });
  }

  function setPassword(password: string) {
    send('set_password', { password });
  }

  // Game actions
  function setRule(rule: string, phaseConfig?: unknown) {
    send('set_rule', { rule, phase_config: phaseConfig });
  }

  function startGame() {
    send('start_game');
  }

  function markCell(row: number, col: number, color: string) {
    send('mark_cell', { row, col, color });
  }

  function unmarkCell(row: number, col: number) {
    send('unmark_cell', { row, col });
  }

  function resetGame() {
    send('reset_game');
  }

  function setName(name: string) {
    send('set_name', { name });
  }

  function setCellText(row: number, col: number, text: string) {
    send('set_cell_text', { row, col, text });
  }

  function setAllCellTexts(texts: string[]) {
    send('set_cell_text', { texts });
  }

  return {
    connect,
    disconnect,
    send,
    createRoom,
    joinRoom,
    leaveRoom,
    listRooms,
    setRole,
    setPassword,
    setRule,
    startGame,
    markCell,
    unmarkCell,
    resetGame,
    setName,
    setCellText,
    setAllCellTexts,
  };
}
