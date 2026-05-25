import { useRef } from 'react'

interface UploadDialogProps {
  open: boolean
  onClose: () => void
  onUpload: (file: File) => void
}

export default function UploadDialog({ open, onClose, onUpload }: UploadDialogProps) {
  const fileRef = useRef<HTMLInputElement>(null)

  if (!open) return null

  const handleUpload = () => {
    const file = fileRef.current?.files?.[0]
    if (file) {
      onUpload(file)
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
          />
        </div>
        <div className="flex justify-end gap-2">
          <button
            onClick={onClose}
            className="px-4 py-2 border rounded hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            onClick={handleUpload}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Upload
          </button>
        </div>
      </div>
    </div>
  )
}
