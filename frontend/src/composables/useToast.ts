import { ref } from 'vue'

export type ToastKind = 'info' | 'warn' | 'error'

export interface Toast {
  id: number
  message: string
  kind: ToastKind
}

const toasts = ref<Toast[]>([])
let nextId = 0

export function useToast() {
  function show(message: string, kind: ToastKind = 'info', ttl = 3000) {
    const id = ++nextId
    toasts.value.push({ id, message, kind })
    setTimeout(() => {
      toasts.value = toasts.value.filter(t => t.id !== id)
    }, ttl)
  }

  return {
    toasts,
    show,
  }
}
