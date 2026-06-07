import { API_BASE } from './config'

export interface ReadingProgressResponse {
  chapter_index: number
  percentage: number
}

export async function saveReadingProgress(documentId: string, chapterIndex: number): Promise<ReadingProgressResponse> {
  const res = await fetch(`${API_BASE}/documents/${documentId}/progress`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ chapter_index: chapterIndex }),
  })
  if (!res.ok) throw new Error('Failed to save progress')
  return res.json()
}

export async function getReadingProgress(documentId: string): Promise<ReadingProgressResponse | null> {
  const res = await fetch(`${API_BASE}/documents/${documentId}/progress`)
  if (!res.ok) return null
  return res.json()
}
