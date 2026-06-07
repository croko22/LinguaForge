import { useQuery, useQueryClient } from '@tanstack/react-query'
import { fetchReadingProgress, saveReadingProgress as saveProgressAPI } from '../services/readingProgress'

export function useReadingProgress(documentId: string) {
  const queryClient = useQueryClient()

  const query = useQuery({
    queryKey: ['reading-progress', documentId],
    queryFn: () => fetchReadingProgress(documentId),
    enabled: !!documentId,
    staleTime: Infinity,
  })

  return {
    savedChapterIndex: query.data ?? null,
    isLoading: query.isLoading,
    saveProgress: (chapterIndex: number) => {
      queryClient.setQueryData(['reading-progress', documentId], chapterIndex)
      saveProgressAPI(documentId, chapterIndex)
    },
  }
}
