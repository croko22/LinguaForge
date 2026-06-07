import { saveReadingProgress, getReadingProgress } from '../api/progress'

const STORAGE_KEY = 'linguaforge-reading-progress'

interface ReadingProgress {
  [documentId: string]: number
}

function loadLocal(): ReadingProgress {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    return raw ? JSON.parse(raw) : {}
  } catch {
    return {}
  }
}

function saveLocal(documentId: string, chapterIndex: number): void {
  try {
    const progress = loadLocal()
    progress[documentId] = chapterIndex
    localStorage.setItem(STORAGE_KEY, JSON.stringify(progress))
  } catch {}
}

export function getReadingProgress(documentId: string): number | null {
  try {
    const progress = loadLocal()
    const saved = progress[documentId]
    return typeof saved === 'number' ? saved : null
  } catch {
    return null
  }
}

export async function getReadingProgressFromDB(documentId: string): Promise<number | null> {
  try {
    const result = await getReadingProgress(documentId)
    return result?.chapter_index ?? null
  } catch {
    return null
  }
}

export async function setReadingProgress(documentId: string, chapterIndex: number): Promise<void> {
  saveLocal(documentId, chapterIndex)
  saveReadingProgress(documentId, chapterIndex).catch(() => {})
}
