import { saveReadingProgress, getReadingProgress as getProgressFromAPI } from '../api/progress'

const STORAGE_KEY = 'linguaforge-reading-page-v2'

interface PageProgress {
  [documentId: string]: number
}

function loadLocal(): PageProgress {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    return raw ? JSON.parse(raw) : {}
  } catch {
    return {}
  }
}

function saveLocal(documentId: string, globalPageIndex: number): void {
  try {
    const progress = loadLocal()
    progress[documentId] = globalPageIndex
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
    const result = await getProgressFromAPI(documentId)
    return result?.chapter_index ?? null
  } catch {
    return null
  }
}

export async function setReadingProgress(
  documentId: string,
  globalPageIndex: number,
  chapterIndex?: number,
): Promise<void> {
  saveLocal(documentId, globalPageIndex)
  if (chapterIndex !== undefined) {
    saveReadingProgress(documentId, chapterIndex).catch(() => {})
  }
}
