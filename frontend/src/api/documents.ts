import type { Chapter, ChapterContent } from '../types'
import { API_BASE } from './config'

export interface DocumentSummary {
  id: string
  title: string
  file_type: string
  file_size: number
  status: string
  language: string
  chapter_count: number
  created_at: string
  cover_url?: string
  error_message?: string
}

export type { Chapter, ChapterContent }

export async function fetchDocuments(): Promise<DocumentSummary[]> {
  const res = await fetch(`${API_BASE}/documents`)
  if (!res.ok) throw new Error('Failed to fetch documents')
  return res.json()
}

export async function fetchDocument(documentId: string): Promise<DocumentSummary> {
  const res = await fetch(`${API_BASE}/documents/${documentId}`)
  if (!res.ok) throw new Error('Failed to fetch document')
  return res.json()
}

export async function uploadDocument(file: File): Promise<DocumentSummary> {
  const formData = new FormData()
  formData.append('file', file)
  const res = await fetch(`${API_BASE}/documents`, {
    method: 'POST',
    body: formData,
  })
  if (!res.ok) throw new Error('Failed to upload document')
  return res.json()
}

export async function fetchChapters(documentId: string): Promise<Chapter[]> {
  const res = await fetch(`${API_BASE}/documents/${documentId}/chapters`)
  if (!res.ok) throw new Error('Failed to fetch chapters')
  return res.json()
}

export async function deleteDocument(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/documents/${id}`, { method: 'DELETE' })
  if (!res.ok) throw new Error('Failed to delete document')
}

export async function fetchChapterContent(documentId: string, chapterIndex: number): Promise<ChapterContent> {
  const res = await fetch(`${API_BASE}/documents/${documentId}/chapters/${chapterIndex}`)
  if (!res.ok) throw new Error('Failed to fetch chapter content')
  return res.json()
}
