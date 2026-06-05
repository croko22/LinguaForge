import { type MouseEvent } from 'react'

interface TextDisplayProps {
  content: string
  onWordClick: (word: string, e: MouseEvent<HTMLSpanElement>) => void
}

function splitIntoWords(text: string): string[] {
  return text.split(/\s+/).filter(Boolean)
}

export default function TextDisplay({ content, onWordClick }: TextDisplayProps) {
  const paragraphs = (content ?? '').split(/\n\n+/).filter(Boolean)

  if (paragraphs.length === 0) {
    return <p className="text-gray-400">No content</p>
  }

  return (
    <div className="leading-relaxed">
      {paragraphs.map((para, pIdx) => (
        <p key={pIdx} className="mb-4">
          {splitIntoWords(para).map((word, wIdx) => (
            <span
              key={`${pIdx}-${wIdx}`}
              onClick={(e) => onWordClick(word, e)}
              className="cursor-pointer hover:bg-amber-100/60 rounded px-0.5 transition-all"
            >
              {word}{' '}
            </span>
          ))}
        </p>
      ))}
    </div>
  )
}
