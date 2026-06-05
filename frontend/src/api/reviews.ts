import { API_BASE } from './config'
import type { DueWord, ReviewCard } from '../types'

export async function fetchDueWords(): Promise<{ words: DueWord[]; count: number }> {
  const res = await fetch(`${API_BASE}/words/due`)
  if (!res.ok) throw new Error('Failed to fetch due words')
  return res.json()
}

export async function fetchDueCount(): Promise<number> {
  const res = await fetch(`${API_BASE}/words/due/count`)
  if (!res.ok) throw new Error('Failed to fetch due count')
  const data = await res.json()
  return data.count
}

export async function submitReview(wordId: string, quality: number): Promise<ReviewCard> {
  const res = await fetch(`${API_BASE}/words/${wordId}/review`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ quality }),
  })
  if (!res.ok) throw new Error('Failed to submit review')
  return res.json()
}
