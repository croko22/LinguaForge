import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Routes, Route, useLocation } from 'react-router-dom'
import LibraryPage from './LibraryPage'
import * as api from '../api/documents'
import type { DocumentSummary } from '../api/documents'

// Mock the API module
vi.mock('../api/documents', () => ({
  fetchDocuments: vi.fn(),
  uploadDocument: vi.fn(),
}))

function LocationDisplay() {
  const location = useLocation()
  return <div data-testid="location">{location.pathname}</div>
}

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <Routes>
          <Route path="/" element={ui} />
          <Route path="/read/:id/:chapterIndex" element={<div>Reader Page</div>} />
        </Routes>
        <LocationDisplay />
      </MemoryRouter>
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

  it('shows upload button', () => {
    vi.mocked(api.fetchDocuments).mockResolvedValue([])

    renderWithProviders(<LibraryPage />)

    expect(screen.getByRole('button', { name: /upload/i })).toBeInTheDocument()
  })

  it('opens upload dialog when upload button is clicked', async () => {
    const user = userEvent.setup()
    vi.mocked(api.fetchDocuments).mockResolvedValue([])

    renderWithProviders(<LibraryPage />)

    await user.click(screen.getByRole('button', { name: /upload/i }))

    expect(screen.getByText(/upload epub/i)).toBeInTheDocument()
  })

  it('closes dialog when cancel is clicked', async () => {
    const user = userEvent.setup()
    vi.mocked(api.fetchDocuments).mockResolvedValue([])

    renderWithProviders(<LibraryPage />)

    // Open dialog
    await user.click(screen.getByRole('button', { name: /upload/i }))
    expect(screen.getByText(/upload epub/i)).toBeInTheDocument()

    // Cancel
    await user.click(screen.getByRole('button', { name: /cancel/i }))

    expect(screen.queryByText(/upload epub/i)).not.toBeInTheDocument()
  })

  it('uploads a file and shows it in the list', async () => {
    const user = userEvent.setup()
    vi.mocked(api.fetchDocuments).mockResolvedValue([])
    vi.mocked(api.uploadDocument).mockResolvedValue({
      id: 'new-1',
      title: 'Uploaded Book',
      file_type: 'epub',
      file_size: 2048,
      status: 'ready',
      language: 'en',
      chapter_count: 3,
      created_at: '2026-06-01T00:00:00Z',
    })

    renderWithProviders(<LibraryPage />)

    // Open dialog
    await user.click(screen.getByRole('button', { name: /upload/i }))

    // Select a file
    const file = new File(['test content'], 'test.epub', { type: 'application/epub+zip' })
    const fileInput = screen.getByLabelText(/file/i)
    await user.upload(fileInput, file)

    // Click upload
    await user.click(screen.getByRole('button', { name: /^upload$/i }))

    // Dialog should close and the new book should appear
    await waitFor(() => {
      expect(screen.queryByText(/upload epub/i)).not.toBeInTheDocument()
    })
    expect(await screen.findByText('Uploaded Book')).toBeInTheDocument()
  })

  it('navigates to reader when clicking a document card', async () => {
    const user = userEvent.setup()
    const mockDocs: DocumentSummary[] = [
      {
        id: 'book-1',
        title: 'Economics 101',
        file_type: 'epub',
        file_size: 2048,
        status: 'ready',
        language: 'en',
        chapter_count: 72,
        created_at: '2026-01-01T00:00:00Z',
      },
    ]
    vi.mocked(api.fetchDocuments).mockResolvedValue(mockDocs)

    renderWithProviders(<LibraryPage />)

    await user.click(await screen.findByText('Economics 101'))

    expect(screen.getByTestId('location').textContent).toBe('/read/book-1/0')
  })
})
