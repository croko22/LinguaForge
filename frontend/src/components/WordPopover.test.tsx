import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import WordPopover from './WordPopover'

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

describe('WordPopover', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockFetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ translation: 'cat' }),
    })
  })

  it('renders the word', () => {
    renderWithProviders(<WordPopover word="gato" onClose={vi.fn()} />)
    expect(screen.getByText('gato')).toBeInTheDocument()
  })

  it('renders listen button', () => {
    renderWithProviders(<WordPopover word="gato" onClose={vi.fn()} />)
    expect(screen.getByRole('button', { name: /listen/i })).toBeInTheDocument()
  })

  it('hides when word is null', () => {
    const { container } = renderWithProviders(<WordPopover word={null} onClose={vi.fn()} />)
    expect(container.firstChild).toBeNull()
  })

  it('calls onClose when clicking outside the popover', async () => {
    const user = userEvent.setup()
    const onClose = vi.fn()
    renderWithProviders(<WordPopover word="gato" onClose={onClose} />)

    await user.click(document.body)
    expect(onClose).toHaveBeenCalled()
  })

  it('shows translation from API', async () => {
    renderWithProviders(<WordPopover word="gato" onClose={vi.fn()} />)

    expect(await screen.findByText(/cat/i)).toBeInTheDocument()
  })

  it('shows loading while translating', () => {
    mockFetch.mockReturnValue(new Promise(() => {}))
    renderWithProviders(<WordPopover word="gato" onClose={vi.fn()} />)

    expect(screen.getByText(/translating/i)).toBeInTheDocument()
  })
})
