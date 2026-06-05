const STORAGE_KEY = 'linguaforge-reading-progress'

interface ReadingProgress {
  [documentId: string]: number
}

function loadProgress(): ReadingProgress {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    return raw ? JSON.parse(raw) : {}
  } catch {
    return {}
  }
}

function saveProgress(documentId: string, chapterIndex: number): void {
  try {
    const progress = loadProgress()
    progress[documentId] = chapterIndex
    localStorage.setItem(STORAGE_KEY, JSON.stringify(progress))
  } catch {
    // localStorage not available, ignore
  }
}

export function getReadingProgress(documentId: string): number | null {
  const progress = loadProgress()
  const saved = progress[documentId]
  return typeof saved === 'number' ? saved : null
}

export function setReadingProgress(documentId: string, chapterIndex: number): void {
  saveProgress(documentId, chapterIndex)
}
