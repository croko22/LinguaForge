import { useQuery } from '@tanstack/react-query'
import { fetchChapters, fetchChapterContent } from '../api/documents'

export function useChapters(documentId: string) {
  return useQuery({
    queryKey: ['documents', documentId, 'chapters'],
    queryFn: () => fetchChapters(documentId),
    enabled: !!documentId,
  })
}

export function useChapterContent(documentId: string, chapterIndex: number) {
  return useQuery({
    queryKey: ['documents', documentId, 'chapters', chapterIndex],
    queryFn: () => fetchChapterContent(documentId, chapterIndex),
    enabled: !!documentId && chapterIndex >= 0,
  })
}
