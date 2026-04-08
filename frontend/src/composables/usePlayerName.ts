import { useRoomStore } from '../stores/room'

export function usePlayerName() {
  const roomStore = useRoomStore()

  function playerName(seat: number): string {
    return roomStore.players.find(p => p.seat === seat)?.nickname || `P${seat}`
  }

  return { playerName }
}
