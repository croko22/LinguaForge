import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import TextDisplay from './TextDisplay'

describe('TextDisplay', () => {
  it('renders text split into words', () => {
    render(<TextDisplay content="Hello world this is a test" onWordClick={() => {}} />)
    expect(screen.getByText('Hello')).toBeInTheDocument()
    expect(screen.getByText('world')).toBeInTheDocument()
    expect(screen.getByText('test')).toBeInTheDocument()
  })

  it('calls onWordClick with the word when clicked', async () => {
    const user = userEvent.setup()
    const onWordClick = vi.fn()
    render(<TextDisplay content="Hello world" onWordClick={onWordClick} />)

    await user.click(screen.getByText('world'))
    expect(onWordClick).toHaveBeenCalledWith('world')
  })

  it('preserves paragraph structure with double newlines', () => {
    const text = 'First paragraph.\n\nSecond paragraph.'
    const { container } = render(<TextDisplay content={text} onWordClick={() => {}} />)
    const paragraphs = container.querySelectorAll('p')
    expect(paragraphs).toHaveLength(2)
    expect(paragraphs[0]).toHaveTextContent('First paragraph.')
    expect(paragraphs[1]).toHaveTextContent('Second paragraph.')
  })
})
