import { API_BASE } from './config'

export interface LanguageCount {
  language: string
  count: number
}

export interface ReviewActivity {
  date: string
  count: number
}

export interface Stats {
  total_documents: number
  total_words: number
  total_chapters: number
  language_counts: LanguageCount[]
  review_activity: ReviewActivity[]
}

export async function fetchStats(): Promise<Stats> {
  const res = await fetch(`${API_BASE}/stats`)
  if (!res.ok) throw new Error('Failed to fetch stats')
  return res.json()
}
