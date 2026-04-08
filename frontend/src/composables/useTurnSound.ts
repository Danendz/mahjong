let ctx: AudioContext | null = null

function playTone(freq: number, startSec: number, durSec: number) {
  if (!ctx) return
  const osc = ctx.createOscillator()
  const gain = ctx.createGain()
  osc.type = 'sine'
  osc.frequency.value = freq
  gain.gain.setValueAtTime(0, ctx.currentTime + startSec)
  gain.gain.linearRampToValueAtTime(0.3, ctx.currentTime + startSec + 0.01)
  gain.gain.linearRampToValueAtTime(0, ctx.currentTime + startSec + durSec)
  osc.connect(gain).connect(ctx.destination)
  osc.start(ctx.currentTime + startSec)
  osc.stop(ctx.currentTime + startSec + durSec)
}

export function useTurnSound() {
  function playTurnSound() {
    try {
      if (!ctx) ctx = new AudioContext()
      if (ctx.state === 'suspended') ctx.resume()
      playTone(523, 0, 0.1) // C5
      playTone(659, 0.12, 0.1) // E5
    } catch {
      // Audio not available — silently ignore
    }
  }

  return { playTurnSound }
}
