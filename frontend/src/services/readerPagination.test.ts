import { describe, it, expect } from 'vitest'
import { paginateReaderContent } from './readerPagination'

describe('paginateReaderContent', () => {
  const metrics = {
    viewportHeight: 720,
    fontSize: 1.125,
    lineHeight: 1.8,
  }

  it('keeps paragraph boundaries when content fits on one page', () => {
    const pages = paginateReaderContent('First paragraph.\n\nSecond paragraph.', metrics)

    expect(pages).toHaveLength(1)
    expect(pages[0]).toContain('First paragraph.')
    expect(pages[0]).toContain('Second paragraph.')
  })

  it('splits long content into multiple pages', () => {
    const longText = Array.from({ length: 420 }, (_, index) => `word${index + 1}`).join(' ')
    const pages = paginateReaderContent(longText, metrics, { minWordsPerPage: 160, maxWordsPerPage: 180 })

    expect(pages.length).toBeGreaterThan(1)
    expect(pages[0].split(/\s+/).length).toBeLessThanOrEqual(180)
  })

  it('returns an empty page placeholder for empty content', () => {
    expect(paginateReaderContent('', metrics)).toEqual([''])
  })
})
