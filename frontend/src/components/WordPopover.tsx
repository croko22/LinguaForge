import { useEffect, useRef } from 'react'
import { useTranslate } from '../hooks/useTranslate'

interface WordPopoverProps {
  word: string | null
  onClose: () => void
  sourceLang?: string
  targetLang?: string
}

export default function WordPopover({ word, onClose, sourceLang = 'es', targetLang = 'en' }: WordPopoverProps) {
  const ref = useRef<HTMLDivElement>(null)
  const { data, isLoading } = useTranslate(word ?? '', sourceLang, targetLang)

  useEffect(() => {
    if (!word) return

    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        onClose()
      }
    }

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
    <div ref={ref} className="absolute z-50 bg-white border rounded-lg shadow-lg p-4 min-w-[200px]">
      <p className="font-semibold text-lg mb-2">{word}</p>
      <p className="text-sm text-gray-500 mb-3">
        {isLoading ? 'Translating...' : data?.translation ?? 'Translation: ...'}
      </p>
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
