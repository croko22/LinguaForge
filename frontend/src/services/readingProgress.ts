import { saveReadingProgress as saveAPI, getReadingProgress as getAPI } from '../api/progress'

export async function fetchReadingProgress(documentId: string): Promise<number | null> {
  try {
    const result = await getAPI(documentId)
    return result?.chapter_index ?? null
  } catch {
    return null
  }
}

const saveTimers = new Map<string, ReturnType<typeof setTimeout>>()

export function saveReadingProgress(documentId: string, chapterIndex: number): void {
  const existing = saveTimers.get(documentId)
  if (existing) clearTimeout(existing)

  saveTimers.set(documentId, setTimeout(() => {
    saveAPI(documentId, chapterIndex).catch(() => {
      if (import.meta.env.DEV) console.warn('Failed to save progress')
    })
    saveTimers.delete(documentId)
  }, 2000))
}
