import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useTranslate } from './useTranslate'

const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  }
}

describe('useTranslate', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('returns translation on success', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ translation: 'hola' }),
    })

    const { result } = renderHook(
      () => useTranslate('hello', 'en', 'es'),
      { wrapper: createWrapper() }
    )

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data?.translation).toBe('hola')
  })

  it('does not fetch when word is empty', () => {
    renderHook(
      () => useTranslate('', 'en', 'es'),
      { wrapper: createWrapper() }
    )

    expect(mockFetch).not.toHaveBeenCalled()
  })
})
