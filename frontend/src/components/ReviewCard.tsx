import { useState, useEffect, useCallback, useRef } from 'react'
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
  const [exiting, setExiting] = useState<'left' | 'right' | null>(null)
  const cardRef = useRef<HTMLDivElement>(null)

  const flip = useCallback(() => {
    if (!disabled && !exiting) setFlipped((f) => !f)
  }, [disabled, exiting])

  useEffect(() => {
    const el = cardRef.current
    if (!el) return

    const handler = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return
      if (e.key === ' ' || e.key === 'Enter') {
        e.preventDefault()
        flip()
      }
    }

    el.addEventListener('keydown', handler)
    return () => el.removeEventListener('keydown', handler)
  }, [flip])

  const handleGrade = useCallback((quality: number) => {
    if (disabled || exiting) return
    setExiting(quality >= 3 ? 'right' : 'left')
    setTimeout(() => onGrade(quality), 200)
  }, [disabled, exiting, onGrade])

  useEffect(() => {
    if (!flipped) return
    const el = cardRef.current
    if (!el) return

    const handler = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return
      const grade = parseInt(e.key, 10)
      if (grade >= 1 && grade <= 4) handleGrade(grade)
    }

    el.addEventListener('keydown', handler)
    return () => el.removeEventListener('keydown', handler)
  }, [flipped, disabled, exiting, handleGrade])

  const exitClass =
    exiting === 'left'
      ? 'opacity-0 -translate-x-40 scale-90 rotate-[-6deg]'
      : exiting === 'right'
        ? 'opacity-0 translate-x-40 scale-90 rotate-[6deg]'
        : ''

  return (
    <div
      ref={cardRef}
      tabIndex={0}
      className="flex flex-col items-center gap-6 outline-none"
    >
      <div
        className={`w-full max-w-lg perspective-[1000px] cursor-pointer transition-all duration-200 ${exitClass}`}
        onClick={flip}
      >
        <div
          className={`relative w-full min-h-[280px] transition-transform duration-500 [transform-style:preserve-3d] ${flipped ? '[transform:rotateY(180deg)]' : ''}`}
        >
          <div className="absolute inset-0 bg-surface border-2 border-border rounded-xl flex flex-col items-center justify-center gap-3 p-6 [backface-visibility:hidden]">
            <p className="text-4xl font-bold text-text text-center break-words max-w-full">{word.word}</p>
            {!flipped && (
              <p className="text-xs text-text-muted mt-2">Click or press Space to reveal</p>
            )}
          </div>
          <div className="absolute inset-0 bg-surface border-2 border-border rounded-xl flex flex-col items-center justify-center gap-6 p-6 [transform:rotateY(180deg)] [backface-visibility:hidden]">
            <p className="text-4xl font-bold text-primary-text text-center break-words max-w-full">
              {word.translation}
            </p>
            <div className="flex gap-4 text-xs text-text-muted">
              <span>✕{word.repetitions}</span>
              <span>{word.interval_days}d</span>
              {word.lapses > 0 && <span className="text-grade-again">{word.lapses} lapses</span>}
            </div>
          </div>
        </div>
      </div>

      {flipped && !exiting && (
        <div className="flex gap-3 animate-in fade-in zoom-in-95 duration-150">
          {grades.map((g) => (
            <button
              key={g.quality}
              onClick={() => handleGrade(g.quality)}
              disabled={disabled}
              className={`px-5 py-2.5 text-text-inverse rounded-lg font-medium transition-all disabled:opacity-50 active:scale-95 ${g.className}`}
            >
              <span className="block text-xs opacity-70">[{g.quality}]</span>
              <span className="text-sm">{g.label}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
