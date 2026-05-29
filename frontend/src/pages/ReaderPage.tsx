import { useState, type MouseEvent } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useChapters, useChapterContent } from '../hooks/useReader'
import TextDisplay from '../components/TextDisplay'
import WordPopover from '../components/WordPopover'
import WordPanel from '../components/WordPanel'

export default function ReaderPage() {
  const { id, chapterIndex: chapterIndexParam } = useParams()
  const navigate = useNavigate()
  const currentIndex = parseInt(chapterIndexParam ?? '0', 10)

  const { data: chapters } = useChapters(id ?? '')
  const { data: chapter } = useChapterContent(id ?? '', currentIndex)

  const [selectedWord, setSelectedWord] = useState<string | null>(null)
  const [popoverPos, setPopoverPos] = useState<{ x: number; y: number } | null>(null)
  const [clickedWords, setClickedWords] = useState<string[]>([])

  const totalChapters = chapters?.length ?? 0
  const hasPrev = currentIndex > 0
  const hasNext = currentIndex < totalChapters - 1

  const goTo = (index: number) => {
    setSelectedWord(null)
    navigate(`/read/${id}/${index}`)
  }

  const handleWordClick = (word: string, e: MouseEvent<HTMLSpanElement>) => {
    const clean = word.replace(/^[^\w]+|[^\w]+$/g, '')
    if (!clean) return
    const rect = (e.target as HTMLElement).getBoundingClientRect()
    setPopoverPos({ x: rect.left + rect.width / 2, y: rect.bottom + 4 })
    setSelectedWord(clean)
    setClickedWords((prev) => (prev.includes(clean) ? prev : [...prev, clean]))
  }

  if (!chapter) {
    return (
      <div className="flex h-screen items-center justify-center">
        <p className="text-gray-500">Loading...</p>
      </div>
    )
  }

  return (
    <div className="flex h-screen">
      {/* Text area */}
      <div className="flex-1 flex flex-col overflow-hidden relative">
        {/* Chapter header */}
        <div className="border-b p-4 flex items-center justify-between">
          <h1 className="text-xl font-semibold">{chapter.chapter_title}</h1>
          <div className="flex items-center gap-2">
            <button
              onClick={() => goTo(currentIndex - 1)}
              disabled={!hasPrev}
              className="px-3 py-1 border rounded disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
            >
              Prev
            </button>
            <select
              value={currentIndex}
              onChange={(e) => goTo(parseInt(e.target.value, 10))}
              className="border rounded px-2 py-1"
            >
              {chapters?.map((ch) => (
                <option key={ch.chapter_index} value={ch.chapter_index}>
                  {ch.chapter_title} content
                </option>
              ))}
            </select>
            <button
              onClick={() => goTo(currentIndex + 1)}
              disabled={!hasNext}
              className="px-3 py-1 border rounded disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
            >
              Next
            </button>
          </div>
        </div>

        {/* Chapter content */}
        <div className="flex-1 overflow-y-auto p-6">
          <TextDisplay content={chapter.content} onWordClick={handleWordClick} />
        </div>

        {/* Word popover */}
        {selectedWord && popoverPos && (
          <WordPopover word={selectedWord} position={popoverPos} onClose={() => { setSelectedWord(null); setPopoverPos(null); }} />
        )}
      </div>

      {/* Word panel */}
      <div className="w-80 border-l p-4 overflow-y-auto">
        <WordPanel words={clickedWords} onClear={() => setClickedWords([])} />
      </div>
    </div>
  )
}
