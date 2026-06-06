export interface ReaderPaginationMetrics {
  viewportHeight: number
  fontSize: number
  lineHeight: number
  chromeHeight?: number
}

export interface ReaderPaginationOptions {
  minWordsPerPage?: number
  maxWordsPerPage?: number
  wordsPerLineFactor?: number
}

function splitParagraphs(content: string): string[] {
  return content
    .split(/\n\s*\n+/)
    .map((paragraph) => paragraph.trim())
    .filter(Boolean)
}

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value))
}

export function paginateReaderContent(
  content: string,
  metrics: ReaderPaginationMetrics,
  options: ReaderPaginationOptions = {},
): string[] {
  const paragraphs = splitParagraphs(content)
  if (paragraphs.length === 0) return ['']

  const chromeHeight = metrics.chromeHeight ?? 96
  const availableHeight = Math.max(0, metrics.viewportHeight - chromeHeight)
  const lineHeightPx = Math.max(1, metrics.fontSize * 16 * metrics.lineHeight)
  const estimatedLines = availableHeight / lineHeightPx
  const wordsPerLineFactor = options.wordsPerLineFactor ?? 12
  const minWordsPerPage = options.minWordsPerPage ?? 160
  const maxWordsPerPage = options.maxWordsPerPage ?? 280
  const wordsPerPage = clamp(
    Math.round(estimatedLines * wordsPerLineFactor),
    minWordsPerPage,
    maxWordsPerPage,
  )

  const pages: string[] = []
  let currentParagraphs: string[] = []
  let currentWordCount = 0

  const flush = () => {
    if (currentParagraphs.length === 0) return
    pages.push(currentParagraphs.join('\n\n'))
    currentParagraphs = []
    currentWordCount = 0
  }

  for (const paragraph of paragraphs) {
    const words = paragraph.split(/\s+/).filter(Boolean)
    if (words.length === 0) continue

    if (currentWordCount > 0 && currentWordCount + words.length > wordsPerPage) {
      flush()
    }

    if (words.length > wordsPerPage) {
      flush()

      for (let index = 0; index < words.length; index += wordsPerPage) {
        pages.push(words.slice(index, index + wordsPerPage).join(' '))
      }

      continue
    }

    currentParagraphs.push(paragraph)
    currentWordCount += words.length

    if (currentWordCount >= wordsPerPage) {
      flush()
    }
  }

  flush()

  return pages.length > 0 ? pages : ['']
}
