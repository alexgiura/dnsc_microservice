import { ref } from 'vue'

/** O singură instanță de meniu deschisă la un moment dat (toate DropdownMenu-urile). */
export const sharedOpenMenuId = ref<string | null>(null)

let seq = 0
export function nextDropdownMenuId(): string {
  return `dm-${++seq}`
}
