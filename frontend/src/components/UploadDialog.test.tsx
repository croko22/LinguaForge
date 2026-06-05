import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import UploadDialog from './UploadDialog'

const mockOnClose = vi.fn()
const mockOnUpload = vi.fn()

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      {ui}
    </QueryClientProvider>,
  )
}

describe('UploadDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders when open is true', () => {
    renderWithProviders(
      <UploadDialog open={true} onClose={mockOnClose} onUpload={mockOnUpload} />
    )
    expect(screen.getByText(/upload book/i)).toBeInTheDocument()
    expect(screen.getByTestId('file-input')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /upload/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
  })

  it('does not render when open is false', () => {
    renderWithProviders(
      <UploadDialog open={false} onClose={mockOnClose} onUpload={mockOnUpload} />
    )
    expect(screen.queryByText(/upload book/i)).not.toBeInTheDocument()
  })

  it('calls onClose when cancel button is clicked', async () => {
    const user = userEvent.setup()
    renderWithProviders(
      <UploadDialog open={true} onClose={mockOnClose} onUpload={mockOnUpload} />
    )
    await user.click(screen.getByRole('button', { name: /cancel/i }))
    expect(mockOnClose).toHaveBeenCalledOnce()
  })
})
