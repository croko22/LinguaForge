import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fetchDueWords, fetchDueCount, submitReview } from '../api/reviews'

export function useDueWords() {
  return useQuery({
    queryKey: ['dueWords'],
    queryFn: fetchDueWords,
  })
}

export function useDueCount() {
  return useQuery({
    queryKey: ['dueCount'],
    queryFn: fetchDueCount,
    refetchInterval: 60_000,
  })
}

export function useSubmitReview() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ wordId, quality }: { wordId: string; quality: number }) =>
      submitReview(wordId, quality),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dueWords'] })
      queryClient.invalidateQueries({ queryKey: ['dueCount'] })
    },
  })
}
