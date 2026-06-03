import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fetchDocuments, uploadDocument } from '../api/documents'

export function useDocuments() {
  const { data, ...rest } = useQuery({
    queryKey: ['documents'],
    queryFn: fetchDocuments,
    refetchInterval: (query) => {
      const docs = query.state.data
      if (!docs) return false
      return docs.some(d => d.status === 'pending' || d.status === 'processing') ? 2000 : false
    },
  })
  return { data, ...rest }
}

export function useUploadDocument() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: uploadDocument,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['documents'] })
    },
  })
}
