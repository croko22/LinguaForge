import { create } from 'zustand'
import { persist } from 'zustand/middleware'

type ReaderTheme = 'light' | 'sepia' | 'dark'

interface ReaderSettings {
  theme: ReaderTheme
  fontSize: number
  lineHeight: number
  setTheme: (theme: ReaderTheme) => void
  setFontSize: (size: number) => void
  setLineHeight: (height: number) => void
}

export const useReaderSettings = create<ReaderSettings>()(
  persist(
    (set) => ({
      theme: 'light',
      fontSize: 1.125,
      lineHeight: 1.8,
      setTheme: (theme) => set({ theme }),
      setFontSize: (fontSize) => set({ fontSize }),
      setLineHeight: (lineHeight) => set({ lineHeight }),
    }),
    { name: 'linguaforge-reader-settings' }
  )
)

export type { ReaderTheme }
