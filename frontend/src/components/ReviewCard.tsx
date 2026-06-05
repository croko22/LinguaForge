import { useState } from 'react'
import type { DueWord } from '../types'

interface ReviewCardProps {
  word: DueWord
  onGrade: (quality: number) => void
  disabled?: boolean
}

const grades = [
  { quality: 1, label: 'Again', className: 'bg-grade-again hover:bg-grade-again-hover' },
  { quality: 2, label: 'Hard', className: 'bg-grade-hard hover:bg-grade-hard-hover' },
  { quality: 3, label: 'Good', className: 'bg-grade-good hover:bg-grade-good-hover' },
  { quality: 4, label: 'Easy', className: 'bg-grade-easy hover:bg-grade-easy-hover' },
]

export default function ReviewCard({ word, onGrade, disabled }: ReviewCardProps) {
  const [flipped, setFlipped] = useState(false)

  return (
    <div className="flex flex-col items-center gap-6">
      <div
        className="w-full max-w-lg perspective-[1000px] cursor-pointer"
        onClick={() => !disabled && setFlipped(!flipped)}
      >
        <div
          className={`relative w-full min-h-[250px] transition-transform duration-500 [transform-style:preserve-3d] ${flipped ? '[transform:rotateY(180deg)]' : ''}`}
        >
          <div className="absolute inset-0 bg-surface border-2 border-border rounded-xl flex items-center justify-center [backface-visibility:hidden]">
            <p className="text-3xl font-bold text-text">{word.word}</p>
          </div>
          <div className="absolute inset-0 bg-surface border-2 border-border rounded-xl flex items-center justify-center [transform:rotateY(180deg)] [backface-visibility:hidden]">
            <p className="text-3xl font-bold text-primary-text">{word.translation}</p>
          </div>
        </div>
      </div>

      {flipped && (
        <div className="flex gap-3">
          {grades.map((g) => (
            <button
              key={g.quality}
              onClick={() => onGrade(g.quality)}
              disabled={disabled}
               className={`px-5 py-2.5 text-text-inverse rounded-lg font-medium transition-colors disabled:opacity-50 ${g.className}`}
            >
              {g.label}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
