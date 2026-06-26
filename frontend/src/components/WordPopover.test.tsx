import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import WordPopover from './WordPopover'

const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)
const pos = { x: 100, y: 200 }

describe('WordPopover', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockFetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ translation: 'cat' }),
    })
  })

  it('renders the word', () => {
    render(<WordPopover word="gato" position={pos} onClose={vi.fn()} />)
    expect(screen.getByText('gato')).toBeInTheDocument()
  })

  it('renders listen button', () => {
    render(<WordPopover word="gato" position={pos} onClose={vi.fn()} />)
    expect(screen.getByRole('button', { name: /listen/i })).toBeInTheDocument()
  })

  it('hides when word is null', () => {
    const { container } = render(<WordPopover word={null} position={pos} onClose={vi.fn()} />)
    expect(container.firstChild).toBeNull()
  })

  it('calls onClose when clicking outside', async () => {
    const user = userEvent.setup()
    const onClose = vi.fn()
    render(<WordPopover word="gato" position={pos} onClose={onClose} />)

    await user.click(document.body)
    expect(onClose).toHaveBeenCalled()
  })

  it('shows translation', () => {
    render(<WordPopover word="gato" position={pos} onClose={vi.fn()} translation="cat" />)
    expect(screen.getByText(/cat/i)).toBeInTheDocument()
  })

  it('shows translation failure', () => {
    render(<WordPopover word="gato" position={pos} onClose={vi.fn()} error />)
    expect(screen.getByText(/translation failed/i)).toBeInTheDocument()
  })
})
