import { useRef, useState } from 'react'

interface UploadDialogProps {
  open: boolean
  onClose: () => void
  onUpload: (file: File) => void
}

export default function UploadDialog({ open, onClose, onUpload }: UploadDialogProps) {
  const fileRef = useRef<HTMLInputElement>(null)
  const [uploading, setUploading] = useState(false)

  if (!open) return null

  const handleUpload = async () => {
    const file = fileRef.current?.files?.[0]
    if (!file || uploading) return
    setUploading(true)
    try {
      await onUpload(file)
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" role="dialog">
      <div className="bg-white rounded-lg p-6 w-full max-w-md">
        <h2 className="text-xl font-semibold mb-4">Upload EPUB</h2>
        <div className="mb-4">
          <label htmlFor="file-input" className="block text-sm font-medium mb-1">
            File
          </label>
          <input
            id="file-input"
            ref={fileRef}
            type="file"
            accept=".epub"
            className="w-full border rounded px-3 py-2"
            disabled={uploading}
          />
        </div>
        <div className="flex justify-end gap-2">
          <button
            onClick={onClose}
            disabled={uploading}
            className="px-4 py-2 border rounded hover:bg-gray-50 disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            onClick={handleUpload}
            disabled={uploading}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
          >
            {uploading ? 'Uploading...' : 'Upload'}
          </button>
        </div>
      </div>
    </div>
  )
}
