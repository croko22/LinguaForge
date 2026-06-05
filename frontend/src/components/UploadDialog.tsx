import { useRef, useState, type DragEvent } from 'react'

interface UploadDialogProps {
  open: boolean
  onClose: () => void
  onUpload: (file: File) => void
}

const ACCEPTED_TYPE = '.epub'
const MAX_SIZE = 50 * 1024 * 1024

function UploadIcon({ className }: { className?: string }) {
  return (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
    </svg>
  )
}

export default function UploadDialog({ open, onClose, onUpload }: UploadDialogProps) {
  const fileRef = useRef<HTMLInputElement>(null)
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [uploading, setUploading] = useState(false)
  const [isDragging, setIsDragging] = useState(false)
  const [fileError, setFileError] = useState<string | null>(null)

  if (!open) return null

  const validateFile = (file: File): string | null => {
    const ext = '.' + file.name.split('.').pop()?.toLowerCase()
    if (ext !== ACCEPTED_TYPE) return 'Only EPUB files are supported'
    if (file.size > MAX_SIZE) return 'File is too large. Maximum size is 50 MB'
    return null
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const error = validateFile(file)
    if (error) {
      setFileError(error)
      setSelectedFile(null)
      return
    }
    setFileError(null)
    setSelectedFile(file)
  }

  const handleDragOver = (e: DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(true)
  }

  const handleDragLeave = (e: DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(false)
  }

  const handleDrop = (e: DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(false)

    const file = e.dataTransfer.files?.[0]
    if (!file) return
    const error = validateFile(file)
    if (error) {
      setFileError(error)
      setSelectedFile(null)
      return
    }
    setFileError(null)
    setSelectedFile(file)
  }

  const handleUpload = async () => {
    if (!selectedFile || uploading) return
    setUploading(true)
    try {
      await onUpload(selectedFile)
      setSelectedFile(null)
      setFileError(null)
    } catch {
      setUploading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" role="dialog">
      <div className="bg-surface rounded-2xl p-6 w-full max-w-md shadow-2xl">
        <h2 className="text-xl font-semibold mb-1">Upload Book</h2>
        <p className="text-sm text-text-secondary mb-4">Add an EPUB to your library</p>

        <div
          className={`border-2 border-dashed rounded-xl p-8 text-center cursor-pointer transition-colors ${
            isDragging
               ? 'border-primary bg-primary-light'
               : fileError
                 ? 'border-red-300 bg-danger-light'
                : 'border-gray-300 hover:border-gray-400'
          }`}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
          onClick={() => fileRef.current?.click()}
        >
          <UploadIcon className="w-10 h-10 mx-auto mb-3 text-text-muted" />
          <p className="text-sm text-text-secondary font-medium">
            {selectedFile ? selectedFile.name : 'Drag & drop your EPUB here'}
          </p>
          <p className="text-xs text-text-muted mt-1">
            {selectedFile ? `${(selectedFile.size / 1024 / 1024).toFixed(1)} MB` : 'or click to browse'}
          </p>
          {fileError && <p className="text-xs text-danger mt-2">{fileError}</p>}
          <input
            ref={fileRef}
            type="file"
            accept=".epub"
            className="hidden"
            onChange={handleFileSelect}
            data-testid="file-input"
          />
        </div>

        <div className="flex justify-end gap-2 mt-4">
          <button
            onClick={onClose}
            disabled={uploading}
            className="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-surface-hover disabled:opacity-50 transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleUpload}
            disabled={!selectedFile || uploading}
            className="px-4 py-2 bg-primary text-text-inverse text-sm font-medium rounded-lg hover:bg-primary-hover disabled:opacity-50 transition-colors"
          >
            {uploading ? 'Uploading...' : 'Upload'}
          </button>
        </div>
      </div>
    </div>
  )
}
