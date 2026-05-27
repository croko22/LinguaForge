import { useQuery } from '@tanstack/react-query'
import { translateWord } from '../api/translate'

export function useTranslate(word: string, sourceLang: string, targetLang: string) {
  return useQuery({
    queryKey: ['translate', word, sourceLang, targetLang],
    queryFn: () => translateWord({ word, source_lang: sourceLang, target_lang: targetLang }),
    enabled: !!word,
  })
}
