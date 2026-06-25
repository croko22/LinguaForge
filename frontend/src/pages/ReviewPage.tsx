import { useState, useEffect, useRef, useCallback } from 'react'
import { useDueWords, useSubmitReview } from '../hooks/useReview'
import ReviewCard from '../components/ReviewCard'

export default function ReviewPage() {
  const { data, isLoading, isError } = useDueWords()
  const submitReview = useSubmitReview()
  const [currentIndex, setCurrentIndex] = useState(0)
  const [correctCount, setCorrectCount] = useState(0)
  const startTime = useRef<number>(0)
  const [elapsed, setElapsed] = useState('0s')

  useEffect(() => {
    startTime.current = Date.now()
  }, [])

  useEffect(() => {
    const id = setInterval(() => {
      const sec = Math.floor((Date.now() - startTime.current) / 1000)
      if (sec < 60) setElapsed(`${sec}s`)
      else setElapsed(`${Math.floor(sec / 60)}m ${sec % 60}s`)
    }, 1000)
    return () => clearInterval(id)
  }, [])

  const words = data?.words ?? []
  const currentWord = words[currentIndex]
  const total = words.length

  const handleGrade = useCallback(
    (quality: number) => {
      if (!currentWord) return
      if (quality >= 3) setCorrectCount((c) => c + 1)

      submitReview.mutate({ wordId: currentWord.id, quality })

      if (currentIndex + 1 >= total) {
        setCurrentIndex(total)
      } else {
        setCurrentIndex((i) => i + 1)
      }
    },
    [currentWord, currentIndex, total, submitReview],
  )

  if (isLoading) {
    return (
      <div className="flex h-[80vh] items-center justify-center">
        <p className="text-text-secondary text-lg animate-pulse">Loading review session...</p>
      </div>
    )
  }

  if (isError) {
    return (
      <div className="flex h-[80vh] items-center justify-center">
        <div className="text-center">
          <p className="text-danger text-lg font-semibold mb-2">Failed to load due words</p>
          <button
            onClick={() => window.location.reload()}
            className="text-primary hover:underline text-sm"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  if (total === 0) {
    return (
      <div className="flex h-[80vh] items-center justify-center">
        <div className="text-center space-y-4">
          <p className="text-6xl">🎉</p>
          <p className="text-xl font-semibold text-text">All caught up!</p>
          <p className="text-text-secondary">No cards due for review. Come back later.</p>
        </div>
      </div>
    )
  }

  if (currentIndex >= total) {
    const accuracy = Math.round((correctCount / total) * 100)
    const raw = elapsed

    return (
      <div className="flex h-[80vh] items-center justify-center">
        <div className="text-center space-y-4 max-w-md">
          <p className="text-6xl">🏁</p>
          <p className="text-xl font-semibold text-text">Session complete!</p>
          <p className="text-text-secondary">{total} cards reviewed in {raw}</p>
          <div className="flex justify-center gap-8 text-sm">
            <div>
              <p className="text-2xl font-bold text-grade-good">{accuracy}%</p>
              <p className="text-text-muted">accuracy</p>
            </div>
            <div>
              <p className="text-2xl font-bold text-text">{correctCount}/{total}</p>
              <p className="text-text-muted">correct (≥Good)</p>
            </div>
            <div>
              <p className="text-2xl font-bold text-text-muted">{raw}</p>
              <p className="text-text-muted">time</p>
            </div>
          </div>
          <button
            onClick={() => {
              setCurrentIndex(0)
              setCorrectCount(0)
              startTime.current = Date.now()
            }}
            className="mt-4 px-6 py-2 bg-primary text-text-inverse rounded-lg hover:bg-primary-hover transition-colors text-sm font-medium"
          >
            Review again
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-6 text-sm">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <div className="w-24 h-2 bg-border rounded-full overflow-hidden">
              <div
                className="h-full bg-primary rounded-full transition-all duration-300"
                style={{ width: `${((currentIndex + 1) / total) * 100}%` }}
              />
            </div>
            <span className="text-text-muted tabular-nums">
              {currentIndex + 1}/{total}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-3 text-text-muted">
          <span>⏱ {elapsed}</span>
          {correctCount > 0 && (
            <span className="text-grade-good">✓ {correctCount}</span>
          )}
        </div>
      </div>

      <ReviewCard
        key={currentWord.id}
        word={currentWord}
        onGrade={handleGrade}
        disabled={submitReview.isPending}
      />

      <p className="text-center text-xs text-text-muted mt-4">
        Space/Enter to flip · 1(Again) 2(Hard) 3(Good) 4(Easy)
      </p>
    </div>
  )
}
