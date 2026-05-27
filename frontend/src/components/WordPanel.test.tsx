import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import WordPanel from './WordPanel'

const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>
  )
}

describe('WordPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockFetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ translation: 'cat' }),
    })
  })

  it('shows empty state when no words clicked', () => {
    renderWithProviders(<WordPanel words={[]} onClear={vi.fn()} />)
    expect(screen.getByText(/click a word/i)).toBeInTheDocument()
  })

  it('shows a list of clicked words', () => {
    renderWithProviders(<WordPanel words={['gato', 'casa', 'sol']} onClear={vi.fn()} />)
    expect(screen.getByText('gato')).toBeInTheDocument()
    expect(screen.getByText('casa')).toBeInTheDocument()
    expect(screen.getByText('sol')).toBeInTheDocument()
  })

  it('shows word count badge', () => {
    renderWithProviders(<WordPanel words={['gato', 'casa']} onClear={vi.fn()} />)
    expect(screen.getByText(/2 words/i)).toBeInTheDocument()
  })

  it('calls onClear when clear button is clicked', async () => {
    const user = userEvent.setup()
    const onClear = vi.fn()
    renderWithProviders(<WordPanel words={['gato']} onClear={onClear} />)

    await user.click(screen.getByRole('button', { name: /clear/i }))
    expect(onClear).toHaveBeenCalledOnce()
  })
})
