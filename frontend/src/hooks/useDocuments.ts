import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fetchDocuments, uploadDocument, DocumentSummary } from '../api/documents'

export function useDocuments() {
  return useQuery({
    queryKey: ['documents'],
    queryFn: fetchDocuments,
  })
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
