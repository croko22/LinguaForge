import { API_BASE } from './config'

export interface SavedWord {
  id: string
  document_id: string
  word: string
  translation: string
  source_lang: string
  target_lang: string
  created_at: string
}

export interface SaveWordParams {
  word: string
  translation: string
  documentId: string
  sourceLang: string
  targetLang: string
}

export async function saveWord({ word, translation, documentId, sourceLang, targetLang }: SaveWordParams): Promise<SavedWord> {
  const res = await fetch(`${API_BASE}/words`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      document_id: documentId,
      word,
      translation,
      source_lang: sourceLang,
      target_lang: targetLang,
    }),
  })
  if (!res.ok) throw new Error('Failed to save word')
  return res.json()
}

export async function loadWords(): Promise<SavedWord[]> {
  const res = await fetch(`${API_BASE}/words`)
  if (!res.ok) throw new Error('Failed to load words')
  return res.json()
}

export async function deleteWord(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/words/${id}`, { method: 'DELETE' })
  if (!res.ok) throw new Error('Failed to delete word')
}
