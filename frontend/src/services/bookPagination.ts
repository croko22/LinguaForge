import { paginateReaderContent, type ReaderPaginationMetrics, type ReaderPaginationOptions } from './readerPagination'

export interface BookPage {
  content: string
  chapterIndex: number
  chapterTitle: string
}

export interface BookPagination {
  pages: BookPage[]
  chapterPageRanges: Map<number, { start: number; end: number }>
}

interface ChapterInput {
  chapter_index: number
  chapter_title: string
  content: string
}

export function paginateBook(
  chapters: ChapterInput[],
  metrics: ReaderPaginationMetrics,
  options: ReaderPaginationOptions = {},
): BookPagination {
  const pages: BookPage[] = []
  const chapterPageRanges = new Map<number, { start: number; end: number }>()

  for (const ch of chapters) {
    const chapterPages = paginateReaderContent(ch.content, metrics, options)
    const start = pages.length
    for (const content of chapterPages) {
      pages.push({ content, chapterIndex: ch.chapter_index, chapterTitle: ch.chapter_title })
    }
    chapterPageRanges.set(ch.chapter_index, { start, end: pages.length - 1 })
  }

  return { pages, chapterPageRanges }
}
