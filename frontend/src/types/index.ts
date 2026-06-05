export interface Document {
  id: string
  title: string
  file_type: string
  file_size: number
  status: string
  language: string
  chapter_count: number
  created_at: string
  cover_url?: string
}

export interface Chapter {
  id: string
  document_id: string
  chapter_index: number
  chapter_title: string
  token_count: number
  created_at: string
}

export interface ChapterContent extends Chapter {
  content: string
}

export interface DueWord {
  id: string
  word: string
  translation: string
  source_lang: string
  target_lang: string
  document_id: string
  status: string
  next_review: string
  ease_factor: number
  interval_days: number
  repetitions: number
  lapses: number
}

export interface ReviewCard {
  id: string
  word_id: string
  status: string
  ease_factor: number
  interval_days: number
  repetitions: number
  lapses: number
  next_review: string
  last_reviewed_at: string | null
  created_at: string
  updated_at: string
}
