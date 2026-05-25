import { useDocuments } from '../hooks/useDocuments'
import type { DocumentSummary } from '../api/documents'

export default function LibraryPage() {
  const { data: documents, isLoading } = useDocuments()

  if (isLoading) {
    return (
      <div className="max-w-6xl mx-auto p-6">
        <h1 className="text-2xl font-bold mb-4">My Library</h1>
        <p className="text-gray-500">Loading...</p>
      </div>
    )
  }

  if (documents && documents.length === 0) {
    return (
      <div className="max-w-6xl mx-auto p-6">
        <h1 className="text-2xl font-bold mb-4">My Library</h1>
        <p className="text-gray-500">No documents yet. Upload your first EPUB.</p>
      </div>
    )
  }

  return (
    <div className="max-w-6xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">My Library</h1>
      <div className="grid gap-4">
        {documents?.map((doc) => (
          <DocumentCard key={doc.id} document={doc} />
        ))}
      </div>
    </div>
  )
}

function DocumentCard({ document }: { document: DocumentSummary }) {
  return (
    <div className="border rounded-lg p-4 hover:shadow-md cursor-pointer">
      <h2 className="font-semibold text-lg">{document.title}</h2>
      <div className="flex gap-2 mt-1 text-sm text-gray-500">
        <span className="uppercase">{document.file_type}</span>
        <span>{document.chapter_count} chapters</span>
        {document.language && <span>{document.language}</span>}
        <span className={statusColor(document.status)}>{document.status}</span>
      </div>
    </div>
  )
}

function statusColor(status: string): string {
  switch (status) {
    case 'ready':
      return 'text-green-600'
    case 'error':
      return 'text-red-600'
    default:
      return 'text-yellow-600'
  }
}
