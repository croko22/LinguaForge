import { useState } from 'react'
import { useDueWords, useSubmitReview } from '../hooks/useReview'
import ReviewCard from '../components/ReviewCard'

export default function ReviewPage() {
  const { data, isLoading, isError } = useDueWords()
  const submitReview = useSubmitReview()
  const [queue, setQueue] = useState<string[]>([])
  const [currentIndex, setCurrentIndex] = useState(0)

  const words = data?.words ?? []
  const currentWord = words[currentIndex]
  const total = words.length

  if (isLoading) {
    return (
      <div className="flex h-[80vh] items-center justify-center">
        <p className="text-gray-500 text-lg">Loading review session...</p>
      </div>
    )
  }

  if (isError) {
    return (
      <div className="flex h-[80vh] items-center justify-center">
        <p className="text-red-500 text-lg">Failed to load due words</p>
      </div>
    )
  }

  if (total === 0) {
    return (
      <div className="flex h-[80vh] items-center justify-center">
        <div className="text-center">
          <p className="text-4xl mb-4">🎉</p>
          <p className="text-xl text-gray-600">No cards due for review!</p>
        </div>
      </div>
    )
  }

  const handleGrade = (quality: number) => {
    if (!currentWord) return
    if (currentIndex + 1 >= total) {
      setCurrentIndex(total)
      submitReview.mutate({ wordId: currentWord.id, quality })
    } else {
      submitReview.mutate({ wordId: currentWord.id, quality })
      setCurrentIndex((i) => i + 1)
    }
  }

  if (currentIndex >= total) {
    return (
      <div className="flex h-[80vh] items-center justify-center">
        <div className="text-center">
          <p className="text-4xl mb-4">🎉</p>
          <p className="text-xl text-gray-600">Session complete! All {total} cards reviewed.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      <div className="text-center mb-6">
        <p className="text-sm text-gray-500">
          {currentIndex + 1} of {total}
        </p>
      </div>

      <ReviewCard
        word={currentWord}
        onGrade={handleGrade}
        disabled={submitReview.isPending}
      />
    </div>
  )
}
