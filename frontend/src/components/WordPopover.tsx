import { useEffect, useRef, useState, useCallback } from 'react'
import { API_BASE } from '../api/config'

interface WordPopoverProps {
  word: string | null
  position: { x: number; y: number }
  onClose: () => void
}

export default function WordPopover({ word, position, onClose }: WordPopoverProps) {
  const ref = useRef<HTMLDivElement>(null)
  const [translation, setTranslation] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(false)

  useEffect(() => {
    if (!word) return
    setLoading(true)
    setError(false)
    setTranslation(null)

    fetch(`${API_BASE}/translate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ word, source_lang: 'en', target_lang: 'es' }),
    })
      .then((res) => {
        if (!res.ok) throw new Error('HTTP ' + res.status)
        return res.json()
      })
      .then((data) => {
        setTranslation(data.translation)
        setLoading(false)
      })
      .catch(() => {
        setError(true)
        setLoading(false)
      })
  }, [word])

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
    utterance.lang = 'en'
    speechSynthesis.speak(utterance)
  }, [word])

  if (!word) return null

  return (
    <div
      ref={ref}
      style={{ left: position.x, top: position.y }}
      className="fixed z-50 bg-surface border rounded-xl shadow-2xl p-4 min-w-[200px] -translate-x-1/2"
    >
      <p className="font-semibold text-base mb-1 text-text">{word}</p>
      <p className="text-sm mb-3">
        {loading ? (
          <span className="text-text-muted italic">Translating...</span>
        ) : error ? (
          <span className="text-danger text-xs">Translation failed</span>
        ) : translation ? (
          <span className="text-primary font-medium">{translation}</span>
        ) : (
          <span className="text-text-muted">Translation: ...</span>
        )}
      </p>
      <div className="flex gap-2">
        <button
          onClick={handleListen}
          title={`Listen to "${word}"`}
          className="flex items-center gap-1.5 px-3 py-1.5 border border-border rounded-lg text-sm text-text-secondary hover:bg-surface-hover transition-all"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" />
          </svg>
          <span className="text-xs font-medium">Listen</span>
        </button>
      </div>
    </div>
  )
}
