import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Routes, Route, useLocation } from 'react-router-dom'
import LibraryPage from './LibraryPage'
import * as api from '../api/documents'
import type { DocumentSummary } from '../api/documents'

vi.mock('../api/documents', () => ({
  fetchDocuments: vi.fn(),
  uploadDocument: vi.fn(),
  deleteDocument: vi.fn(),
}))

vi.mock('../api/config', () => ({
  API_BASE: '/api',
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
          <Route path="/read/:id" element={<div>Reader Page</div>} />
          <Route path="/read/:id/:chapterIndex" element={<div>Reader Page</div>} />
        </Routes>
        <LocationDisplay />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

const baseDoc: DocumentSummary = {
  id: '1',
  title: 'Test Book',
  file_type: 'epub',
  file_size: 1024,
  status: 'ready',
  language: 'en',
  chapter_count: 5,
  created_at: '2026-01-01T00:00:00Z',
}

describe('LibraryPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders loading skeleton with pulse animation', () => {
    vi.mocked(api.fetchDocuments).mockReturnValue(new Promise(() => {}))
    const { container } = renderWithProviders(<LibraryPage />)
    expect(screen.getByText('My Library')).toBeInTheDocument()
    expect(container.querySelector('.animate-pulse')).toBeInTheDocument()
  })

  it('renders empty state when no documents', async () => {
    vi.mocked(api.fetchDocuments).mockResolvedValue([])
    renderWithProviders(<LibraryPage />)
    expect(await screen.findByText('No books yet')).toBeInTheDocument()
    expect(screen.getByText(/upload your first epub/i)).toBeInTheDocument()
  })

  it('renders error state with retry button', async () => {
    vi.mocked(api.fetchDocuments).mockRejectedValue(new Error('fail'))
    renderWithProviders(<LibraryPage />)
    expect(await screen.findByText('Failed to load your library')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /try again/i })).toBeInTheDocument()
  })

  it('retry button refetches documents', async () => {
    const user = userEvent.setup()
    vi.mocked(api.fetchDocuments)
      .mockRejectedValueOnce(new Error('fail'))
      .mockResolvedValueOnce([{ ...baseDoc, id: '2', title: 'Retried Book' }])

    renderWithProviders(<LibraryPage />)
    await user.click(await screen.findByRole('button', { name: /try again/i }))
    expect(await screen.findByRole('heading', { name: /Retried Book/ })).toBeInTheDocument()
  })

  it('renders document cards with cover images', async () => {
    const docWithCover: DocumentSummary = {
      ...baseDoc,
      cover_url: 'covers/abc.jpg',
    }
    vi.mocked(api.fetchDocuments).mockResolvedValue([docWithCover])

    renderWithProviders(<LibraryPage />)

    expect(await screen.findByText('Test Book')).toBeInTheDocument()
    const img = screen.getByRole('img')
    expect(img).toHaveAttribute('src', expect.stringContaining('/api/documents/1/cover'))
    expect(img).toHaveAttribute('alt', 'Test Book')
  })

  it('renders generated cover when no cover image', async () => {
    vi.mocked(api.fetchDocuments).mockResolvedValue([baseDoc])

    renderWithProviders(<LibraryPage />)

    expect(await screen.findByRole('heading', { name: /Test Book/ })).toBeInTheDocument()
    expect(screen.queryByRole('img')).not.toBeInTheDocument()
    expect(screen.getAllByText('en').length).toBe(2)
  })

  it('navigates to reader when clicking a card', async () => {
    const user = userEvent.setup()
    vi.mocked(api.fetchDocuments).mockResolvedValue([baseDoc])

    renderWithProviders(<LibraryPage />)
    await user.click(await screen.findByRole('heading', { name: /Test Book/ }))
    expect(screen.getByTestId('location').textContent).toBe('/read/1')
  })

  it('opens upload dialog when upload button is clicked', async () => {
    const user = userEvent.setup()
    vi.mocked(api.fetchDocuments).mockResolvedValue([])

    renderWithProviders(<LibraryPage />)
    await user.click(await screen.findByRole('button', { name: /upload/i }))

    expect(screen.getByText(/upload book/i)).toBeInTheDocument()
  })

  it('closes dialog when cancel is clicked', async () => {
    const user = userEvent.setup()
    vi.mocked(api.fetchDocuments).mockResolvedValue([])

    renderWithProviders(<LibraryPage />)

    await user.click(await screen.findByRole('button', { name: /upload/i }))
    expect(screen.getByText(/upload book/i)).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: /cancel/i }))
    expect(screen.queryByText(/upload book/i)).not.toBeInTheDocument()
  })

  it('uploads a file and shows it in the list', async () => {
    const user = userEvent.setup()
    vi.mocked(api.fetchDocuments).mockResolvedValue([])
    vi.mocked(api.uploadDocument).mockResolvedValue({
      ...baseDoc,
      id: 'new-1',
      title: 'Uploaded Book',
      chapter_count: 3,
    })

    renderWithProviders(<LibraryPage />)

    await user.click(await screen.findByRole('button', { name: /upload/i }))

    const file = new File(['test'], 'test.epub', { type: 'application/epub+zip' })
    const fileInput = screen.getByTestId('file-input')
    await user.upload(fileInput, file)

    await user.click(screen.getByRole('button', { name: /^upload$/i }))

    await waitFor(() => {
      expect(screen.queryByText(/upload book/i)).not.toBeInTheDocument()
    })
    expect(await screen.findByRole('heading', { name: /Uploaded Book/ })).toBeInTheDocument()
  })

  describe('status color-coding', () => {
    it('shows green badge for ready status', async () => {
      vi.mocked(api.fetchDocuments).mockResolvedValue([{ ...baseDoc, status: 'ready' }])
      renderWithProviders(<LibraryPage />)
      const badge = await screen.findByText('ready')
      expect(badge.className).toContain('text-green-700')
      expect(badge.className).toContain('bg-green-50')
    })

    it('shows red badge for error status', async () => {
      vi.mocked(api.fetchDocuments).mockResolvedValue([{ ...baseDoc, status: 'error' }])
      renderWithProviders(<LibraryPage />)
      const badge = await screen.findByText('error')
      expect(badge.className).toContain('text-danger-text')
      expect(badge.className).toContain('bg-danger-light')
    })

    it('shows amber badge for processing status', async () => {
      vi.mocked(api.fetchDocuments).mockResolvedValue([{ ...baseDoc, status: 'processing' }])
      renderWithProviders(<LibraryPage />)
      const badge = await screen.findByText('processing')
      expect(badge.className).toContain('text-amber-700')
      expect(badge.className).toContain('bg-amber-50')
    })

    it('shows amber badge for pending status', async () => {
      vi.mocked(api.fetchDocuments).mockResolvedValue([{ ...baseDoc, status: 'pending' }])
      renderWithProviders(<LibraryPage />)
      const badge = await screen.findByText('pending')
      expect(badge.className).toContain('text-amber-700')
      expect(badge.className).toContain('bg-amber-50')
    })
  })

  describe('metadata display', () => {
    it('shows file type as uppercase text', async () => {
      vi.mocked(api.fetchDocuments).mockResolvedValue([{ ...baseDoc, file_type: 'epub' }])
      renderWithProviders(<LibraryPage />)
      expect(await screen.findByText('epub')).toBeInTheDocument()
    })

    it('shows chapter count', async () => {
      vi.mocked(api.fetchDocuments).mockResolvedValue([{ ...baseDoc, chapter_count: 12 }])
      renderWithProviders(<LibraryPage />)
      expect(await screen.findByText('12 chapters')).toBeInTheDocument()
    })

    it('shows singular chapter count for single chapter', async () => {
      vi.mocked(api.fetchDocuments).mockResolvedValue([{ ...baseDoc, chapter_count: 1 }])
      renderWithProviders(<LibraryPage />)
      expect(await screen.findByText('1 chapter')).toBeInTheDocument()
    })

    it('shows language when provided', async () => {
      vi.mocked(api.fetchDocuments).mockResolvedValue([{ ...baseDoc, language: 'fr' }])
      renderWithProviders(<LibraryPage />)
      const langs = await screen.findAllByText('fr')
      expect(langs.length).toBe(2)
    })
  })

  describe('view toggle', () => {
    it('renders grid and list toggle buttons', async () => {
      vi.mocked(api.fetchDocuments).mockResolvedValue([baseDoc])
      renderWithProviders(<LibraryPage />)
      await screen.findByRole('heading', { name: /Test Book/ })

      const gridBtn = screen.getByTitle('Grid view')
      const listBtn = screen.getByTitle('List view')
      expect(gridBtn).toBeInTheDocument()
      expect(listBtn).toBeInTheDocument()
    })

    it('clicking list toggle shows table view', async () => {
      const user = userEvent.setup()
      vi.mocked(api.fetchDocuments).mockResolvedValue([
        { ...baseDoc, id: '1', title: 'List Book', file_type: 'pdf', chapter_count: 3, language: 'en', status: 'ready' },
      ])
      renderWithProviders(<LibraryPage />)
      await screen.findByRole('heading', { name: /List Book/ })

      await user.click(screen.getByTitle('List view'))

      expect(screen.getByText('Book')).toBeInTheDocument()
      expect(screen.getByText('Type')).toBeInTheDocument()
      expect(screen.getByText('Chapters')).toBeInTheDocument()
      expect(screen.getByText('Language')).toBeInTheDocument()
      expect(screen.getByText('Status')).toBeInTheDocument()
    })

    it('clicking grid toggle shows grid view after list', async () => {
      const user = userEvent.setup()
      vi.mocked(api.fetchDocuments).mockResolvedValue([baseDoc])
      renderWithProviders(<LibraryPage />)
      await screen.findByRole('heading', { name: /Test Book/ })

      await user.click(screen.getByTitle('List view'))
      expect(screen.getByText('Book')).toBeInTheDocument()

      await user.click(screen.getByTitle('Grid view'))
      expect(screen.queryByText('Type')).not.toBeInTheDocument()
    })

    it('list view shows document metadata', async () => {
      const user = userEvent.setup()
      vi.mocked(api.fetchDocuments).mockResolvedValue([
        { ...baseDoc, id: '1', title: 'Meta Book', file_type: 'mobi', chapter_count: 7, language: 'de', status: 'processing' },
      ])
      renderWithProviders(<LibraryPage />)
      await screen.findByRole('heading', { name: /Meta Book/ })

      await user.click(screen.getByTitle('List view'))

      expect(screen.getByText('mobi')).toBeInTheDocument()
      expect(screen.getByText('7')).toBeInTheDocument()
      expect(screen.getByText('de')).toBeInTheDocument()
      expect(screen.getByText('processing')).toBeInTheDocument()
    })
  })
})
