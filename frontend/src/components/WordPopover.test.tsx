import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import WordPopover from './WordPopover'

describe('WordPopover', () => {
  it('renders the word and action buttons', () => {
    render(<WordPopover word="gato" onClose={vi.fn()} />)
    expect(screen.getByText('gato')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /translate/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /listen/i })).toBeInTheDocument()
  })

  it('hides when word is null', () => {
    const { container } = render(<WordPopover word={null} onClose={vi.fn()} />)
    expect(container.firstChild).toBeNull()
  })

  it('calls onClose when clicking outside the popover', async () => {
    const user = userEvent.setup()
    const onClose = vi.fn()
    render(<WordPopover word="gato" onClose={onClose} />)

    await user.click(document.body)
    expect(onClose).toHaveBeenCalled()
  })

  it('renders translation placeholder text', () => {
    render(<WordPopover word="gato" onClose={vi.fn()} />)
    expect(screen.getByText(/translation/i)).toBeInTheDocument()
  })
})
