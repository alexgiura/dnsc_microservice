import { apiFetch } from '@/api/client'

export interface TagDto {
  id: string
  value: string
  sort_order: number
}

export async function listTags(): Promise<TagDto[]> {
  const res = await apiFetch('/api/tags')
  if (!res.ok) {
    const t = await res.text()
    throw new Error(t || `HTTP ${res.status}`)
  }
  const data = (await res.json()) as { tags: TagDto[] }
  return Array.isArray(data.tags) ? data.tags : []
}
