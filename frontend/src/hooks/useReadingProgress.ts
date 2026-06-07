import { saveReadingProgress as saveProgressAPI, getReadingProgress as getProgressFromAPI } from '../api/progress'

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

let saveTimeout: ReturnType<typeof setTimeout> | null = null

function saveProgressToDB(documentId: string, chapterIndex: number): void {
  if (saveTimeout) clearTimeout(saveTimeout)
  saveTimeout = setTimeout(() => {
    saveProgressAPI(documentId, chapterIndex).catch(() => {
      if (import.meta.env.DEV) console.warn('Failed to save progress to DB')
    })
  }, 3000)
}

export function setReadingProgress(
  documentId: string,
  globalPageIndex: number,
  chapterIndex?: number,
): void {
  saveLocal(documentId, globalPageIndex)
  if (chapterIndex !== undefined) {
    saveProgressToDB(documentId, chapterIndex)
  }
}
