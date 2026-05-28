import { useState, useEffect } from 'react'
import { API_BASE } from '../api/config'

interface WordPanelProps {
  words: string[]
  onClear: () => void
}

function WordItem({ word }: { word: string }) {
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
    <li className="text-sm border-b pb-1">
      <span className="font-medium">{word}</span>
      <p className="text-xs text-gray-400">{translation ?? 'Translating...'}</p>
    </li>
  )
}

export default function WordPanel({ words, onClear }: WordPanelProps) {
  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <h2 className="font-semibold text-sm uppercase tracking-wide text-gray-500">
          Words <span className="ml-1 text-xs bg-gray-200 rounded-full px-2 py-0.5">{words.length} words</span>
        </h2>
        {words.length > 0 && (
          <button onClick={onClear} className="text-xs text-red-500 hover:text-red-700">
            Clear
          </button>
        )}
      </div>
      {words.length === 0 ? (
        <p className="text-sm text-gray-400">Click a word to add it here</p>
      ) : (
        <ul className="space-y-2">
          {words.map((word, idx) => (
            <WordItem key={`${word}-${idx}`} word={word} />
          ))}
        </ul>
      )}
    </div>
  )
}
