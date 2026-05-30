export interface Document {
  id: string
  title: string
  file_type: string
  file_size: number
  status: string
  language: string
  chapter_count: number
  created_at: string
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
