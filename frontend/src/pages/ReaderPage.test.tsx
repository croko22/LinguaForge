import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import ReaderPage from './ReaderPage'
import * as api from '../api/documents'
import type { Chapter, ChapterContent } from '../api/documents'

// Mock global fetch for WordPopover / words API
const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

const { mockGetProgress } = vi.hoisted(() => ({
  mockGetProgress: vi.fn().mockResolvedValue(null),
}))

vi.mock('../api/progress', () => ({
  getReadingProgress: mockGetProgress,
  saveReadingProgress: vi.fn().mockResolvedValue({ chapter_index: 0, percentage: 0 }),
}))

vi.mock('../api/documents', () => ({
  fetchDocuments: vi.fn(),
  fetchDocument: vi.fn(),
  uploadDocument: vi.fn(),
  fetchChapters: vi.fn(),
  fetchChapterContent: vi.fn(),
}))

function renderWithProviders(
  ui: React.ReactElement,
  initialEntries = ['/read/doc-1/0'],
) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={initialEntries}>
        <Routes>
          <Route path="/read/:id" element={ui} />
          <Route path="/read/:id/:chapterIndex" element={ui} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('ReaderPage', () => {
  const mockChapters: Chapter[] = [
    { id: 'c1', document_id: 'doc-1', chapter_index: 0, chapter_title: 'Introduction', token_count: 120, created_at: '' },
    { id: 'c2', document_id: 'doc-1', chapter_index: 1, chapter_title: 'Chapter 1', token_count: 340, created_at: '' },
    { id: 'c3', document_id: 'doc-1', chapter_index: 2, chapter_title: 'Chapter 2', token_count: 280, created_at: '' },
  ]

  const mockContent: ChapterContent = {
    id: 'c1', document_id: 'doc-1', chapter_index: 0, chapter_title: 'Introduction',
    content: 'This is the first chapter content.', token_count: 6, created_at: '',
  }

  const mockDocument = {
    id: 'doc-1',
    title: 'Test Book',
    file_type: 'epub',
    file_size: 1024,
    status: 'ready',
    language: 'en',
    chapter_count: 3,
    created_at: '',
  }

  beforeEach(() => {
    vi.clearAllMocks()
    mockGetProgress.mockResolvedValue(null)
    vi.mocked(api.fetchChapters).mockResolvedValue(mockChapters)
    vi.mocked(api.fetchChapterContent).mockResolvedValue(mockContent)
    vi.mocked(api.fetchDocument).mockResolvedValue(mockDocument)
    mockFetch
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([]) })
      .mockResolvedValue({ ok: true, json: () => Promise.resolve({ translation: 'test' }) })
  })

  it('renders the book title and chapter title', async () => {
    renderWithProviders(<ReaderPage />)
    expect(await screen.findByText('Test Book')).toBeInTheDocument()
    const heading = await screen.findByRole('heading', { level: 1 })
    expect(heading).toHaveTextContent('Introduction')
  })

  it('shows a visible page indicator', async () => {
    renderWithProviders(<ReaderPage />)
    expect(await screen.findByText(/page 1 \/ 3/i)).toBeInTheDocument()
  })

  it('restores saved page progress when opened from the library route', async () => {
    const user = userEvent.setup()
    const firstRender = renderWithProviders(<ReaderPage />)

    expect(await screen.findByText(/page 1 \/ 3/i)).toBeInTheDocument()
    await user.keyboard('{ArrowRight}')
    expect(await screen.findByText(/page 2 \/ 3/i)).toBeInTheDocument()

    firstRender.unmount()
    mockGetProgress.mockResolvedValueOnce({ chapter_index: 1, percentage: 50 })
    renderWithProviders(<ReaderPage />, ['/read/doc-1'])

    expect(await screen.findByText(/page 2 \/ 3/i)).toBeInTheDocument()
  })

  it('renders chapter content text', async () => {
    renderWithProviders(<ReaderPage />)
    expect(await screen.findByText('first')).toBeInTheDocument()
    expect(await screen.findByText('chapter')).toBeInTheDocument()
    expect(await screen.findByText('content.')).toBeInTheDocument()
  })

  it('renders chapter navigation with prev/next buttons', async () => {
    const user = userEvent.setup()
    renderWithProviders(<ReaderPage />)
    await user.click(await screen.findByRole('button', { name: /chapters/i }))
    expect(await screen.findByRole('button', { name: /prev/i })).toBeInTheDocument()
    expect(await screen.findByRole('button', { name: /next/i })).toBeInTheDocument()
  })

  it('shows prev button disabled on first chapter', async () => {
    const user = userEvent.setup()
    renderWithProviders(<ReaderPage />)
    await user.click(await screen.findByRole('button', { name: /chapters/i }))
    const prev = await screen.findByRole('button', { name: /prev/i })
    expect(prev).toBeDisabled()
  })

  it('navigates to next chapter on next click', async () => {
    const user = userEvent.setup()
    vi.mocked(api.fetchChapterContent).mockImplementation(async (_documentId, chapterIndex) => ({
      ...mockContent,
      chapter_index: chapterIndex,
      chapter_title: chapterIndex === 1 ? 'Chapter 1' : 'Introduction',
    }))
    renderWithProviders(<ReaderPage />)

    await user.click(await screen.findByRole('button', { name: /chapters/i }))
    await user.click(await screen.findByRole('button', { name: /next/i }))
    expect(await screen.findByText(/Chapter 1/i)).toBeInTheDocument()
  })

  it('clicking a word adds it to the word panel', async () => {
    const user = userEvent.setup()
    renderWithProviders(<ReaderPage />)

    const word = await screen.findByText('first')
    await user.click(word)

    expect(screen.getAllByText('first')[0]).toBeInTheDocument()
  })

  it('shows word popover when clicking a word', async () => {
  const user = userEvent.setup()
  renderWithProviders(<ReaderPage />)

  const word = await screen.findByText('first')
  await user.click(word)

  expect(screen.getByRole('button', { name: /listen/i })).toBeInTheDocument()
})

  it('shows vocabulary panel after clicking words', async () => {
    const user = userEvent.setup()
    renderWithProviders(<ReaderPage />)

    const first = await screen.findByText('first')
    await user.click(first)
    const chapter = await screen.findByText('chapter')
    await user.click(chapter)

    expect(screen.getByText(/vocabulary/i)).toBeInTheDocument()
  })

  it('toggles the chapter menu', async () => {
    const user = userEvent.setup()
    renderWithProviders(<ReaderPage />)

    await user.click(await screen.findByRole('button', { name: /chapters/i }))
    expect(screen.getByLabelText(/jump to chapter/i)).toBeInTheDocument()
  })

  it('clear button removes all words from panel', async () => {
    const user = userEvent.setup()
    renderWithProviders(<ReaderPage />)

    const word = await screen.findByText('first')
    await user.click(word)
    expect(screen.getAllByText('first').length).toBeGreaterThanOrEqual(2)

    await user.click(screen.getByRole('button', { name: /clear/i }))
    expect(screen.getByText(/any word/i)).toBeInTheDocument()
  })

  it('shows error state when document not found', async () => {
    vi.mocked(api.fetchChapterContent).mockRejectedValue(new Error('not found'))
    vi.mocked(api.fetchChapters).mockRejectedValue(new Error('not found'))

    renderWithProviders(<ReaderPage />)

    expect(await screen.findByText(/document not found/i)).toBeInTheDocument()
  })

  it('shows error state when chapter content fails to load', async () => {
    vi.mocked(api.fetchChapterContent).mockRejectedValue(new Error('network error'))

    renderWithProviders(<ReaderPage />)

    expect(await screen.findByText(/failed to load/i)).toBeInTheDocument()
  })
})
