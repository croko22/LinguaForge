import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { fetchDocuments, uploadDocument, deleteDocument } from '../api/documents'

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

export function useDeleteDocument() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteDocument,
    onSuccess: () => {
      toast.success('Book deleted')
      queryClient.invalidateQueries({ queryKey: ['documents'] })
    },
    onError: () => {
      toast.error('Failed to delete book')
    },
  })
}

export function useUploadDocument() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: uploadDocument,
    onSuccess: () => {
      toast.success('Book uploaded successfully')
      queryClient.invalidateQueries({ queryKey: ['documents'] })
    },
    onError: () => {
      toast.error('Failed to upload book')
    },
  })
}
