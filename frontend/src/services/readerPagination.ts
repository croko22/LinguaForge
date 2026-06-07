export interface ReaderPaginationMetrics {
  viewportHeight: number
  viewportWidth?: number
  fontSize: number
  lineHeight: number
  chromeHeight?: number
}

export interface ReaderPaginationOptions {
  minWordsPerPage?: number
  maxWordsPerPage?: number
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
  const maxChWidth = 68
  const charWidthPx = metrics.fontSize * 16 * 0.6
  const containerWidth = metrics.viewportWidth
    ? Math.min(metrics.viewportWidth, maxChWidth * charWidthPx)
    : maxChWidth * charWidthPx
  const avgWordLength = 6
  const wordsPerLine = Math.max(1, Math.floor(containerWidth / (avgWordLength * charWidthPx)))
  // If viewport is too small to measure, use generous defaults so pages aren't tiny
  const estimatedLines = availableHeight > 50
    ? availableHeight / lineHeightPx
    : Math.max(20, (window.innerHeight ?? 800) / lineHeightPx)
  const defaultMinWordsPerPage = 30
  const defaultMaxWordsPerPage = 1200
  const minWordsPerPage = options.minWordsPerPage ?? defaultMinWordsPerPage
  const maxWordsPerPage = options.maxWordsPerPage ?? defaultMaxWordsPerPage
  const wordsPerPage = clamp(
    Math.round(estimatedLines * wordsPerLine),
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
