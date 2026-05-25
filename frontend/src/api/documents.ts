const API_BASE = '/api'

export interface DocumentSummary {
  id: string
  title: string
  file_type: string
  file_size: number
  status: string
  language: string
  chapter_count: number
  created_at: string
}

export async function fetchDocuments(): Promise<DocumentSummary[]> {
  const res = await fetch(`${API_BASE}/documents`)
  if (!res.ok) throw new Error('Failed to fetch documents')
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
