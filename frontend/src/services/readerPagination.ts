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
  const estimatedLines = availableHeight / lineHeightPx
  const maxChWidth = 68
  const charWidthPx = metrics.fontSize * 16 * 0.6
  const containerWidth = metrics.viewportWidth
    ? Math.min(metrics.viewportWidth, maxChWidth * charWidthPx)
    : maxChWidth * charWidthPx
  const avgWordLength = 6
  const wordsPerLine = Math.max(1, Math.floor(containerWidth / (avgWordLength * charWidthPx)))
  const minWordsPerPage = options.minWordsPerPage ?? 30
  const maxWordsPerPage = options.maxWordsPerPage ?? 1200
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
