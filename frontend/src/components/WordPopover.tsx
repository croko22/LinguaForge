import { useEffect, useRef, useState, useCallback } from 'react'
import { API_BASE } from '../api/config'
import { playWordAudio } from '../api/tts'
import { useLanguageSettings } from '../store/languageSettings'

interface WordPopoverProps {
  word: string | null
  position: { x: number; y: number }
  onClose: () => void
  language?: string
}

export default function WordPopover({ word, position, onClose, language }: WordPopoverProps) {
  const { sourceLang, targetLang } = useLanguageSettings()
  const ref = useRef<HTMLDivElement>(null)
  const [result, setResult] = useState<{
    word: string | null
    translation: string | null
    error: boolean
  }>({ word: null, translation: null, error: false })

  const isCurrentResult = result.word === word
  const translation = isCurrentResult ? result.translation : null
  const error = isCurrentResult ? result.error : false
  const loading = !!word && !isCurrentResult

  useEffect(() => {
    if (!word) return
    let cancelled = false

    fetch(`${API_BASE}/translate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ word, source_lang: language ?? sourceLang, target_lang: targetLang }),
    })
      .then((res) => {
        if (!res.ok) throw new Error('HTTP ' + res.status)
        return res.json()
      })
      .then((data) => {
        if (!cancelled) {
          setResult({ word, translation: data.translation, error: false })
        }
      })
      .catch(() => {
        if (!cancelled) {
          setResult({ word, translation: null, error: true })
        }
      })

    return () => {
      cancelled = true
    }
  }, [word, language, sourceLang, targetLang])

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
    playWordAudio(word, language ?? sourceLang)
  }, [word, language, sourceLang])

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
