import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import WordPanel from './WordPanel'

const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

describe('WordPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockFetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ translation: 'cat' }),
    })
  })

  it('shows empty state', () => {
    render(<WordPanel words={[]} onClear={vi.fn()} />)
    expect(screen.getByText(/click a word/i)).toBeInTheDocument()
  })

  it('shows list of clicked words', () => {
    render(<WordPanel words={['gato', 'casa', 'sol']} onClear={vi.fn()} />)
    expect(screen.getByText('gato')).toBeInTheDocument()
    expect(screen.getByText('casa')).toBeInTheDocument()
    expect(screen.getByText('sol')).toBeInTheDocument()
  })

  it('shows word count badge', () => {
    render(<WordPanel words={['gato', 'casa']} onClear={vi.fn()} />)
    expect(screen.getByText(/2 words/i)).toBeInTheDocument()
  })

  it('calls onClear when clear button clicked', async () => {
    const user = userEvent.setup()
    const onClear = vi.fn()
    render(<WordPanel words={['gato']} onClear={onClear} />)

    await user.click(screen.getByRole('button', { name: /clear/i }))
    expect(onClear).toHaveBeenCalledOnce()
  })
})
