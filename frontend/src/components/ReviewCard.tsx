import { useState } from 'react'
import type { DueWord } from '../types'

interface ReviewCardProps {
  word: DueWord
  onGrade: (quality: number) => void
  disabled?: boolean
}

const grades = [
  { quality: 1, label: 'Again', className: 'bg-red-500 hover:bg-red-600' },
  { quality: 2, label: 'Hard', className: 'bg-orange-500 hover:bg-orange-600' },
  { quality: 3, label: 'Good', className: 'bg-green-500 hover:bg-green-600' },
  { quality: 4, label: 'Easy', className: 'bg-blue-500 hover:bg-blue-600' },
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
          <div className="absolute inset-0 bg-white border-2 border-gray-200 rounded-xl flex items-center justify-center [backface-visibility:hidden]">
            <p className="text-3xl font-bold text-gray-800">{word.word}</p>
          </div>
          <div className="absolute inset-0 bg-white border-2 border-gray-200 rounded-xl flex items-center justify-center [transform:rotateY(180deg)] [backface-visibility:hidden]">
            <p className="text-3xl font-bold text-blue-700">{word.translation}</p>
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
              className={`px-5 py-2.5 text-white rounded-lg font-medium transition-colors disabled:opacity-50 ${g.className}`}
            >
              {g.label}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
