import { useEffect, useRef, useCallback } from 'react'
import { useTranslate } from '../hooks/useTranslate'

interface WordPopoverProps {
  word: string | null
  position: { x: number; y: number }
  onClose: () => void
  sourceLang?: string
  targetLang?: string
}

export default function WordPopover({ word, position, onClose, sourceLang = 'es', targetLang = 'en' }: WordPopoverProps) {
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

  const handleListen = useCallback(() => {
    if (!word) return
    const utterance = new SpeechSynthesisUtterance(word)
    utterance.lang = sourceLang
    speechSynthesis.speak(utterance)
  }, [word, sourceLang])

  if (!word) return null

  return (
    <div
      ref={ref}
      style={{ left: position.x, top: position.y }}
      className="fixed z-50 bg-white border rounded-lg shadow-lg p-4 min-w-[200px] -translate-x-1/2"
    >
      {/* Word */}
      <p className="font-semibold text-lg mb-1">{word}</p>

      {/* Translation */}
      <p className="text-sm mb-3">
        {isLoading ? (
          <span className="text-gray-400 italic">Translating...</span>
        ) : data?.translation ? (
          <span className="text-blue-700 font-medium">{data.translation}</span>
        ) : (
          <span className="text-gray-400">Translation: ...</span>
        )}
      </p>

      {/* Actions */}
      <div className="flex gap-2">
        <button
          onClick={handleListen}
          title={`Listen to "${word}"`}
          className="flex items-center gap-1 px-3 py-1.5 border rounded text-sm hover:bg-gray-50 transition-colors"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" />
          </svg>
          Listen
        </button>
      </div>
    </div>
  )
}
