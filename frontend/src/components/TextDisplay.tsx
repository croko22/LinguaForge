interface TextDisplayProps {
  content: string
  onWordClick: (word: string) => void
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
    <div className="leading-relaxed text-lg">
      {paragraphs.map((para, pIdx) => (
        <p key={pIdx} className="mb-4">
          {splitIntoWords(para).map((word, wIdx) => (
            <span
              key={`${pIdx}-${wIdx}`}
              onClick={() => onWordClick(word)}
              className="cursor-pointer hover:bg-yellow-100 rounded px-0.5 transition-colors"
            >
              {word}{' '}
            </span>
          ))}
        </p>
      ))}
    </div>
  )
}
