const API_BASE = '/api'

export interface TranslateRequest {
  word: string
  source_lang: string
  target_lang: string
}

export interface TranslateResponse {
  translation: string
}

export async function translateWord(req: TranslateRequest): Promise<TranslateResponse> {
  const res = await fetch(`${API_BASE}/translate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  })
  if (!res.ok) throw new Error('Translation failed')
  return res.json()
}
