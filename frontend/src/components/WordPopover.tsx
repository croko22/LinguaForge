import { useEffect, useRef } from 'react'

interface WordPopoverProps {
  word: string | null
  onClose: () => void
}

export default function WordPopover({ word, onClose }: WordPopoverProps) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!word) return

    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        onClose()
      }
    }

    // Delay to prevent the click that opened it from closing it
    const timer = setTimeout(() => {
      document.addEventListener('mousedown', handleClick)
    }, 0)

    return () => {
      clearTimeout(timer)
      document.removeEventListener('mousedown', handleClick)
    }
  }, [word, onClose])

  if (!word) return null

  return (
    <div
      ref={ref}
      className="absolute z-50 bg-white border rounded-lg shadow-lg p-4 min-w-[200px]"
    >
      <p className="font-semibold text-lg mb-2">{word}</p>
      <p className="text-sm text-gray-500 mb-3">Translation: ...</p>
      <div className="flex gap-2">
        <button className="px-3 py-1 bg-blue-600 text-white rounded text-sm hover:bg-blue-700">
          Translate
        </button>
        <button className="px-3 py-1 border rounded text-sm hover:bg-gray-50">
          Listen
        </button>
      </div>
    </div>
  )
}
