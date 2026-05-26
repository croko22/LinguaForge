import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import ReaderPage from './ReaderPage'
import * as api from '../api/documents'
import type { Chapter, ChapterContent } from '../api/documents'

vi.mock('../api/documents', () => ({
  fetchDocuments: vi.fn(),
  uploadDocument: vi.fn(),
  fetchChapters: vi.fn(),
  fetchChapterContent: vi.fn(),
}))

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={['/read/doc-1/0']}>
        <Routes>
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

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(api.fetchChapters).mockResolvedValue(mockChapters)
    vi.mocked(api.fetchChapterContent).mockResolvedValue(mockContent)
  })

  it('renders the document title as heading', async () => {
    renderWithProviders(<ReaderPage />)
    expect(await screen.findByText('Introduction')).toBeInTheDocument()
  })

  it('renders chapter content text', async () => {
    renderWithProviders(<ReaderPage />)
    expect(await screen.findByText(/this is the first chapter content/i)).toBeInTheDocument()
  })

  it('renders chapter navigation with prev/next buttons', async () => {
    renderWithProviders(<ReaderPage />)
    expect(await screen.findByRole('button', { name: /next/i })).toBeInTheDocument()
  })

  it('shows prev button disabled on first chapter', async () => {
    renderWithProviders(<ReaderPage />)
    const prev = await screen.findByRole('button', { name: /prev/i })
    expect(prev).toBeDisabled()
  })

  it('navigates to next chapter on next click', async () => {
    const user = userEvent.setup()
    vi.mocked(api.fetchChapterContent).mockResolvedValue(mockContent)
    renderWithProviders(<ReaderPage />)

    await user.click(await screen.findByRole('button', { name: /next/i }))
    expect(await screen.findByText(/Chapter 1 content/i)).toBeInTheDocument()
  })
})
