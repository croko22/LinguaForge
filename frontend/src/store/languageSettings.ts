import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface LanguageSettings {
  sourceLang: string
  targetLang: string
  setSourceLang: (lang: string) => void
  setTargetLang: (lang: string) => void
}

export const useLanguageSettings = create<LanguageSettings>()(
  persist(
    (set) => ({
      sourceLang: 'en',
      targetLang: 'es',
      setSourceLang: (sourceLang) => set({ sourceLang }),
      setTargetLang: (targetLang) => set({ targetLang }),
    }),
    { name: 'linguaforge-language-settings' }
  )
)
