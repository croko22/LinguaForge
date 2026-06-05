import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { useDocuments, useUploadDocument, useDeleteDocument } from "../hooks/useDocuments";
import UploadDialog from "../components/UploadDialog";
import { API_BASE } from "../api/config";
import type { DocumentSummary } from "../api/documents";

export default function LibraryPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { data: documents, isLoading, isError, refetch } = useDocuments();
  const uploadMutation = useUploadDocument();
  const [isUploadOpen, setIsUploadOpen] = useState(false);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');

  const handleUpload = async (file: File) => {
    const result = await uploadMutation.mutateAsync(file);
    queryClient.setQueryData<DocumentSummary[]>(["documents"], (old) => [
      ...(old ?? []),
      result,
    ]);
    setIsUploadOpen(false);
  };

  const handleDocumentClick = (id: string, status: string) => {
    if (status === 'ready') {
      navigate(`/read/${id}/0`)
    }
  }

  return (
    <div className="bg-gray-50 min-h-screen">
      <div className="max-w-7xl mx-auto px-6 py-8">
        <header className="flex justify-between items-center mb-8">
          <h1 className="text-2xl font-bold text-gray-900">My Library</h1>
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-2">
              <button
                onClick={() => setViewMode('grid')}
                className={`p-2 rounded-lg transition-colors ${viewMode === 'grid' ? 'bg-gray-200 text-gray-900' : 'text-gray-400 hover:text-gray-600'}`}
                title="Grid view"
              >
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M4 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H4zm0 8a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H4zm6-6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zm0 8a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" clipRule="evenodd" />
                </svg>
              </button>
              <button
                onClick={() => setViewMode('list')}
                className={`p-2 rounded-lg transition-colors ${viewMode === 'list' ? 'bg-gray-200 text-gray-900' : 'text-gray-400 hover:text-gray-600'}`}
                title="List view"
              >
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z" clipRule="evenodd" />
                </svg>
              </button>
            </div>
            {!isUploadOpen && (
              <button
                onClick={() => setIsUploadOpen(true)}
                className="bg-blue-600 text-white px-5 py-2.5 rounded-xl hover:bg-blue-700 transition-colors font-medium flex items-center gap-2"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
                Upload
              </button>
            )}
          </div>
        </header>

        {isLoading ? (
          <LoadingSkeleton viewMode={viewMode} />
        ) : isError ? (
          <ErrorState onRetry={() => refetch()} />
        ) : !documents?.length ? (
          <EmptyState onUpload={() => setIsUploadOpen(true)} />
        ) : viewMode === 'grid' ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {documents.map((doc) => (
              <DocumentCard
                key={doc.id}
                document={doc}
                onClick={() => handleDocumentClick(doc.id, doc.status)}
              />
            ))}
          </div>
        ) : (
          <DocumentList documents={documents} onDocumentClick={(id, status) => handleDocumentClick(id, status)} />
        )}
      </div>

      <UploadDialog
        open={isUploadOpen}
        onClose={() => setIsUploadOpen(false)}
        onUpload={handleUpload}
      />
    </div>
  );
}

function LoadingSkeleton({ viewMode = 'grid' }: { viewMode?: 'grid' | 'list' }) {
  if (viewMode === 'list') {
    return (
      <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden animate-pulse">
        <div className="p-4 space-y-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="flex items-center gap-4">
              <div className="w-10 h-14 bg-gray-200 rounded" />
              <div className="flex-1 space-y-2">
                <div className="h-4 bg-gray-200 rounded w-1/3" />
                <div className="h-3 bg-gray-100 rounded w-1/4" />
              </div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
      {[1, 2, 3, 4].map((i) => (
        <div key={i} className="bg-white rounded-xl overflow-hidden border border-gray-100 animate-pulse">
          <div className="aspect-[3/4] bg-gray-200" />
          <div className="p-4 space-y-3">
            <div className="h-4 bg-gray-200 rounded w-3/4" />
            <div className="h-4 bg-gray-200 rounded w-1/2" />
            <div className="h-3 bg-gray-100 rounded w-1/3" />
          </div>
        </div>
      ))}
    </div>
  );
}

function ErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <div className="text-center py-16">
      <div className="text-5xl mb-4">⚠️</div>
      <h2 className="text-lg font-medium text-gray-900 mb-2">Failed to load your library</h2>
      <p className="text-gray-500 mb-6">Could not fetch your documents. Please try again.</p>
      <button
        onClick={onRetry}
        className="bg-blue-600 text-white px-5 py-2.5 rounded-xl hover:bg-blue-700 transition-colors font-medium"
      >
        Try Again
      </button>
    </div>
  );
}

function EmptyState({ onUpload }: { onUpload: () => void }) {
  return (
    <div className="text-center py-16">
      <div className="text-6xl mb-4">📚</div>
      <h2 className="text-lg font-medium text-gray-900 mb-2">No books yet</h2>
      <p className="text-gray-500 mb-6">Upload your first EPUB to start reading</p>
      <button
        onClick={onUpload}
        className="bg-blue-600 text-white px-5 py-2.5 rounded-xl hover:bg-blue-700 transition-colors font-medium"
      >
        Upload your first book
      </button>
    </div>
  );
}

function DocumentList({ documents, onDocumentClick }: { documents: DocumentSummary[]; onDocumentClick: (id: string, status: string) => void }) {
  const deleteMutation = useDeleteDocument()
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null)

  const handleDelete = (e: React.MouseEvent, docId: string) => {
    e.stopPropagation()
    if (confirmDeleteId === docId) {
      deleteMutation.mutate(docId)
      setConfirmDeleteId(null)
    } else {
      setConfirmDeleteId(docId)
      setTimeout(() => setConfirmDeleteId(null), 3000)
    }
  }
  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
      <table className="w-full">
        <thead>
          <tr className="border-b border-gray-100 text-left text-xs font-medium text-gray-400 uppercase tracking-wider">
            <th className="px-6 py-3">Book</th>
            <th className="px-6 py-3">Type</th>
            <th className="px-6 py-3">Chapters</th>
            <th className="px-6 py-3">Language</th>
            <th className="px-6 py-3">Status</th>
            <th className="px-6 py-3 w-16"></th>
          </tr>
        </thead>
        <tbody>
          {documents.map((doc) => (
            <tr
              key={doc.id}
              onClick={() => onDocumentClick(doc.id, doc.status)}
              className={`border-b border-gray-50 hover:bg-gray-50 cursor-pointer transition-colors ${doc.status === 'error' ? 'opacity-60' : ''}`}
            >
              <td className="px-6 py-4">
                <div className="flex items-center gap-3">
                  {doc.cover_url ? (
                    <img
                      src={`${API_BASE}/documents/${doc.id}/cover`}
                      alt=""
                      className="w-10 h-14 object-cover rounded"
                    />
                  ) : (
                    <div className="w-10 h-14 rounded flex items-center justify-center text-[10px] text-white font-bold leading-tight text-center p-1"
                         style={{ background: 'linear-gradient(135deg, #1e3a5f 0%, #2d5986 50%, #1a365d 100%)' }}>
                      {doc.title}
                    </div>
                  )}
                  <span className="font-medium text-gray-900">{doc.title}</span>
                </div>
              </td>
              <td className="px-6 py-4 text-sm text-gray-500 uppercase">{doc.file_type}</td>
              <td className="px-6 py-4 text-sm text-gray-500">{doc.chapter_count}</td>
              <td className="px-6 py-4 text-sm text-gray-500">{doc.language || '-'}</td>
              <td className="px-6 py-4">
                {doc.status === 'error' ? (
                  <span className="group/error relative inline-block">
                    <span className="inline-block px-2 py-0.5 rounded-full text-xs font-medium text-red-700 bg-red-50 cursor-help">
                      error
                    </span>
                    <span className="absolute bottom-full left-0 mb-2 px-3 py-1.5 bg-gray-900 text-white text-xs rounded-lg whitespace-nowrap opacity-0 group-hover/error:opacity-100 transition-opacity pointer-events-none z-10">
                      {doc.error_message || 'Unknown error'}
                    </span>
                  </span>
                ) : (
                  <span className={`inline-block px-2 py-0.5 rounded-full text-xs font-medium ${
                    doc.status === 'ready' ? 'text-green-700 bg-green-50' :
                    'text-amber-700 bg-amber-50'
                  }`}>
                    {doc.status}
                  </span>
                )}
              </td>
              <td className="px-6 py-4">
                <button
                  onClick={(e) => handleDelete(e, doc.id)}
                  className={`transition-colors ${
                    confirmDeleteId === doc.id ? 'text-red-500' : 'text-gray-400 hover:text-red-500'
                  }`}
                  title={confirmDeleteId === doc.id ? 'Click again to confirm' : 'Delete'}
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-2a1 1 0 00-1 1v3m-4 0h14" />
                  </svg>
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function DocumentCard({
  document: doc,
  onClick,
}: {
  document: DocumentSummary;
  onClick: () => void;
}) {
  const deleteMutation = useDeleteDocument()
  const [confirmDelete, setConfirmDelete] = useState(false)
  const coverUrl = doc.cover_url
    ? `${API_BASE}/documents/${doc.id}/cover`
    : null;

  const isProcessing = doc.status === "processing" || doc.status === "pending";

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (confirmDelete) {
      deleteMutation.mutate(doc.id)
      setConfirmDelete(false)
    } else {
      setConfirmDelete(true)
      setTimeout(() => setConfirmDelete(false), 3000)
    }
  }

  const isError = doc.status === "error"
  const isClickable = doc.status === "ready"

  return (
    <div
      className={`group bg-white rounded-xl shadow-sm transition-all duration-200 overflow-hidden border border-gray-100 relative ${
        isClickable
          ? 'cursor-pointer hover:shadow-lg hover:border-gray-200 hover:-translate-y-0.5'
          : isError
          ? 'opacity-60'
          : 'cursor-default'
      }`}
      onClick={isClickable ? onClick : undefined}
    >
      {coverUrl ? (
        <img
          src={coverUrl}
          alt={doc.title}
          className="aspect-[3/4] w-full object-cover rounded-t-xl"
        />
      ) : (
        <div className="aspect-[3/4] w-full flex flex-col items-center justify-center p-4 text-center"
             style={{ background: 'linear-gradient(135deg, #1e3a5f 0%, #2d5986 50%, #1a365d 100%)' }}>
          <span className="text-white font-bold text-lg leading-tight line-clamp-3">
            {doc.title}
          </span>
          {doc.language && (
            <span className="text-white/60 text-xs mt-2 uppercase tracking-wider">{doc.language}</span>
          )}
        </div>
      )}
      <button
        onClick={handleDelete}
        className={`absolute top-2 right-2 p-1.5 rounded-full transition-all ${
          confirmDelete
            ? 'bg-red-500 text-white opacity-100'
            : 'bg-white/80 text-gray-400 hover:text-red-500 opacity-0 group-hover:opacity-100'
        }`}
        title={confirmDelete ? 'Click again to confirm' : 'Delete'}
      >
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-2a1 1 0 00-1 1v3m-4 0h14" />
        </svg>
      </button>
      <div className="p-4">
        <h3 className="font-semibold text-base line-clamp-2 text-gray-900 mb-2" title={doc.title}>
          {doc.title}
        </h3>
        <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-gray-500 mb-2">
          <span className="text-[11px] uppercase tracking-wider font-medium text-gray-400">{doc.file_type}</span>
          <span className="text-gray-300">·</span>
          <span>{doc.chapter_count} {doc.chapter_count === 1 ? "chapter" : "chapters"}</span>
          {doc.language && (
            <>
              <span className="text-gray-300">·</span>
              <span>{doc.language}</span>
            </>
          )}
        </div>
        {doc.status === 'error' ? (
          <span className="group/error relative inline-block">
            <span className="inline-block px-2 py-0.5 rounded-full text-xs font-medium text-red-700 bg-red-50 cursor-help">
              error
            </span>
            <span className="absolute bottom-full left-0 mb-2 px-3 py-1.5 bg-gray-900 text-white text-xs rounded-lg whitespace-nowrap opacity-0 group-hover/error:opacity-100 transition-opacity pointer-events-none z-10">
              {doc.error_message || 'Unknown error'}
            </span>
          </span>
        ) : (
          <span
            className={`inline-block px-2 py-0.5 rounded-full text-xs font-medium ${
              doc.status === "ready"
                ? "text-green-700 bg-green-50"
                : "text-amber-700 bg-amber-50"
            } ${isProcessing ? "animate-pulse" : ""}`}
          >
            {doc.status}
          </span>
        )}
      </div>
    </div>
  );
}
