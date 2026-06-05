import { useState, useEffect } from 'react'
import { API_BASE } from '../api/config'

interface WordPanelProps {
  words: string[]
  onClear: () => void
}

function WordItem({ word, count }: { word: string; count: number }) {
  const [translation, setTranslation] = useState<string | null>(null)

  useEffect(() => {
    fetch(`${API_BASE}/translate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ word, source_lang: 'en', target_lang: 'es' }),
    })
      .then((r) => r.json())
      .then((data) => setTranslation(data.translation))
      .catch(() => {})
  }, [word])

  return (
    <li className="group flex items-center justify-between px-3 py-2 rounded-lg hover:bg-surface-hover transition-colors cursor-default">
      <div className="flex items-center gap-2 min-w-0">
        <span className="font-semibold text-sm truncate">{word}</span>
        {count > 1 && (
          <span className="text-[10px] bg-gray-200 text-gray-600 px-1.5 py-0.5 rounded-full leading-none font-medium">
            ×{count}
          </span>
        )}
      </div>
      <span className="text-xs text-text-muted italic truncate ml-2 shrink-0">
        {translation ?? '...'}
      </span>
    </li>
  )
}

export default function WordPanel({ words, onClear }: WordPanelProps) {
  const unique = words.reduce<{ word: string; count: number }[]>((acc, w) => {
    const existing = acc.find(item => item.word === w)
    if (existing) existing.count++
    else acc.push({ word: w, count: 1 })
    return acc
  }, [])

  return (
    <div>
      <div className="flex items-center justify-between mb-2 px-1">
        <h2 className="text-xs font-semibold uppercase tracking-wider text-text-muted">
          Vocabulary
          <span className="ml-1.5 text-[10px] bg-gray-200 text-text-secondary rounded-full px-1.5 py-0.5 leading-none">{words.length}</span>
        </h2>
        {words.length > 0 && (
          <button onClick={onClear} className="text-[11px] text-text-muted hover:text-danger transition-colors">
            Clear all
          </button>
        )}
      </div>
      {words.length === 0 ? (
        <p className="text-sm text-text-muted px-1">Click any word to look it up</p>
      ) : (
        <ul className="space-y-0.5">
          {unique.map(({ word, count }) => (
            <WordItem key={word} word={word} count={count} />
          ))}
        </ul>
      )}
    </div>
  )
}
