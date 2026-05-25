import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import LibraryPage from './LibraryPage'
import * as api from '../api/documents'
import type { DocumentSummary } from '../api/documents'

// Mock the API module
vi.mock('../api/documents', () => ({
  fetchDocuments: vi.fn(),
  uploadDocument: vi.fn(),
}))

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>{ui}</MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('LibraryPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows empty state when there are no documents', async () => {
    vi.mocked(api.fetchDocuments).mockResolvedValue([])

    renderWithProviders(<LibraryPage />)

    // Should show empty state message
    expect(await screen.findByText(/no documents/i)).toBeInTheDocument()
    expect(await screen.findByText(/upload your first/i)).toBeInTheDocument()
  })

  it('shows loading state while fetching documents', () => {
    // Don't resolve the promise — keep it pending
    vi.mocked(api.fetchDocuments).mockReturnValue(new Promise(() => {}))

    renderWithProviders(<LibraryPage />)

    expect(screen.getByText(/loading/i)).toBeInTheDocument()
  })

  it('renders document list when API returns data', async () => {
    const mockDocs: DocumentSummary[] = [
      {
        id: '1',
        title: 'Test Book',
        file_type: 'epub',
        file_size: 1024,
        status: 'ready',
        language: 'en',
        chapter_count: 5,
        created_at: '2026-01-01T00:00:00Z',
      },
    ]
    vi.mocked(api.fetchDocuments).mockResolvedValue(mockDocs)

    renderWithProviders(<LibraryPage />)

    expect(await screen.findByText('Test Book')).toBeInTheDocument()
    expect(screen.getByText(/5 chapters/i)).toBeInTheDocument()
    expect(screen.getByText(/epub/i)).toBeInTheDocument()
  })
})
