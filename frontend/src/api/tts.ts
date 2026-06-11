import { API_BASE } from './config'

let voiceLoadPromise: Promise<SpeechSynthesisVoice[]> | null = null

function ensureVoices(): Promise<SpeechSynthesisVoice[]> {
  if (voiceLoadPromise) return voiceLoadPromise
  voiceLoadPromise = new Promise((resolve) => {
    if (!('speechSynthesis' in window)) {
      resolve([])
      return
    }
    const voices = speechSynthesis.getVoices()
    if (voices.length > 0) {
      resolve(voices)
      return
    }
    speechSynthesis.onvoiceschanged = () => {
      resolve(speechSynthesis.getVoices())
    }
    setTimeout(() => resolve(speechSynthesis.getVoices() ?? []), 3000)
  })
  return voiceLoadPromise
}

function mapLangToBCP47(lang: string): string {
  const map: Record<string, string> = {
    en: 'en-US',
    es: 'es-ES',
    fr: 'fr-FR',
    de: 'de-DE',
    it: 'it-IT',
    pt: 'pt-BR',
    ja: 'ja-JP',
    ko: 'ko-KR',
    zh: 'zh-CN',
    ru: 'ru-RU',
    ar: 'ar-SA',
    nl: 'nl-NL',
    pl: 'pl-PL',
    sv: 'sv-SE',
    da: 'da-DK',
    fi: 'fi-FI',
    nb: 'nb-NO',
    tr: 'tr-TR',
  }
  return map[lang] ?? 'en-US'
}

async function speakViaBrowser(word: string, language: string): Promise<boolean> {
  if (!('speechSynthesis' in window)) return false

  speechSynthesis.cancel()

  const voices = await ensureVoices()
  const bcp47 = mapLangToBCP47(language)
  const matchedVoice = voices.find((v) => v.lang.startsWith(bcp47))

  return new Promise((resolve) => {
    const utterance = new SpeechSynthesisUtterance(word)
    utterance.lang = bcp47
    utterance.rate = 0.8
    if (matchedVoice) utterance.voice = matchedVoice

    const timeout = setTimeout(() => {
      speechSynthesis.cancel()
      resolve(false)
    }, 5000)

    utterance.onend = () => {
      clearTimeout(timeout)
      resolve(true)
    }
    utterance.onerror = () => {
      clearTimeout(timeout)
      resolve(false)
    }
    speechSynthesis.speak(utterance)
  })
}

async function speakViaAPI(word: string, language: string): Promise<boolean> {
  try {
    const response = await fetch(
      `${API_BASE}/tts?word=${encodeURIComponent(word)}&lang=${encodeURIComponent(language)}`,
    )
    if (!response.ok) return false
    const blob = await response.blob()
    const url = URL.createObjectURL(blob)
    return new Promise((resolve) => {
      const audio = new Audio(url)
      const timeout = setTimeout(() => {
        audio.pause()
        URL.revokeObjectURL(url)
        resolve(false)
      }, 10000)
      audio.onended = () => {
        clearTimeout(timeout)
        URL.revokeObjectURL(url)
        resolve(true)
      }
      audio.onerror = () => {
        clearTimeout(timeout)
        URL.revokeObjectURL(url)
        resolve(false)
      }
      audio.play().catch(() => {
        clearTimeout(timeout)
        URL.revokeObjectURL(url)
        resolve(false)
      })
    })
  } catch {
    return false
  }
}

export async function playWordAudio(word: string, language: string): Promise<void> {
  // Always try browser SpeechSynthesis first — works offline, no network
  const ok = await speakViaBrowser(word, language)
  if (ok) return
  // Only hit the API if browser TTS wasn't available or failed silently
  await speakViaAPI(word, language)
}
