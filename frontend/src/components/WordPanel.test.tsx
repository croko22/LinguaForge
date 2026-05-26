import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import WordPanel from './WordPanel'

describe('WordPanel', () => {
  it('shows empty state when no words clicked', () => {
    render(<WordPanel words={[]} onClear={vi.fn()} />)
    expect(screen.getByText(/click a word/i)).toBeInTheDocument()
  })

  it('shows a list of clicked words', () => {
    render(<WordPanel words={['gato', 'casa', 'sol']} onClear={vi.fn()} />)
    expect(screen.getByText('gato')).toBeInTheDocument()
    expect(screen.getByText('casa')).toBeInTheDocument()
    expect(screen.getByText('sol')).toBeInTheDocument()
  })

  it('shows word count badge', () => {
    render(<WordPanel words={['gato', 'casa']} onClear={vi.fn()} />)
    expect(screen.getByText(/2 words/i)).toBeInTheDocument()
  })

  it('calls onClear when clear button is clicked', async () => {
    const user = userEvent.setup()
    const onClear = vi.fn()
    render(<WordPanel words={['gato']} onClear={onClear} />)

    await user.click(screen.getByRole('button', { name: /clear/i }))
    expect(onClear).toHaveBeenCalledOnce()
  })
})
