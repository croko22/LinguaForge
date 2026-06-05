import { useEffect, useRef } from 'react'
import { useReaderSettings } from '../store/readerSettings'

interface SettingsPanelProps {
  onClose: () => void
}

export default function SettingsPanel({ onClose }: SettingsPanelProps) {
  const { theme, fontSize, lineHeight, setTheme, setFontSize, setLineHeight } = useReaderSettings()
  const panelRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (panelRef.current && !panelRef.current.contains(e.target as Node)) {
        onClose()
      }
    }
    function handleEscape(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }

    const timer = setTimeout(() => {
      document.addEventListener('mousedown', handleClickOutside)
      document.addEventListener('keydown', handleEscape)
    }, 0)

    return () => {
      clearTimeout(timer)
      document.removeEventListener('mousedown', handleClickOutside)
      document.removeEventListener('keydown', handleEscape)
    }
  }, [onClose])

  const fontSteps = [1, 1.125, 1.25, 1.375, 1.5]
  const fontLabels = ['S', 'M', 'L', 'XL', 'XXL']
  const currentFontIndex = fontSteps.indexOf(fontSize)

  return (
    <div
      ref={panelRef}
      className="absolute right-0 top-full mt-2 w-72 bg-white border rounded-xl shadow-xl p-4 z-50"
      onClick={(e) => e.stopPropagation()}
    >
      <div className="space-y-4">
        <div>
          <label className="block text-xs text-gray-500 uppercase tracking-wide font-semibold mb-2">Theme</label>
          <div className="flex gap-2">
            {(['light', 'sepia', 'dark'] as const).map((t) => (
              <button
                key={t}
                onClick={() => setTheme(t)}
                className={`flex items-center gap-1.5 flex-1 py-2 text-xs font-medium rounded-lg border transition-all
                  ${theme === t
                    ? 'border-blue-500 bg-blue-50 text-blue-700 shadow-sm'
                    : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50 text-gray-500'
                  }`}
              >
                <span
                  className="inline-block w-4 h-4 rounded border shrink-0"
                  style={{
                    backgroundColor: t === 'light' ? '#ffffff' : t === 'sepia' ? '#f4ecd8' : '#1a1a2e',
                    borderColor: t === 'light' ? '#d1d5db' : t === 'sepia' ? '#d4c5a9' : '#334155',
                  }}
                />
                {t === 'light' ? 'Light' : t === 'sepia' ? 'Sepia' : 'Dark'}
              </button>
            ))}
          </div>
        </div>

        <div>
          <label className="block text-xs text-gray-500 uppercase tracking-wide font-semibold mb-2">Font Size</label>
          <div className="flex items-center gap-2">
            <button
              onClick={() => currentFontIndex > 0 && setFontSize(fontSteps[currentFontIndex - 1])}
              disabled={currentFontIndex <= 0}
              className="w-8 h-8 flex items-center justify-center border rounded-lg disabled:opacity-30 hover:bg-gray-100 transition-colors text-gray-600"
            >−</button>
            <span className="text-sm font-semibold min-w-[2rem] text-center text-gray-700">{fontLabels[currentFontIndex]}</span>
            <button
              onClick={() => currentFontIndex < fontSteps.length - 1 && setFontSize(fontSteps[currentFontIndex + 1])}
              disabled={currentFontIndex >= fontSteps.length - 1}
              className="w-8 h-8 flex items-center justify-center border rounded-lg disabled:opacity-30 hover:bg-gray-100 transition-colors text-gray-600"
            >+</button>
          </div>
        </div>

        <div>
          <label className="block text-xs text-gray-500 uppercase tracking-wide font-semibold mb-2">Spacing</label>
          <div className="flex gap-2">
            {([1.6, 1.8, 2.0] as const).map((h) => (
              <button
                key={h}
                onClick={() => setLineHeight(h)}
                className={`flex-1 py-1.5 text-xs font-medium rounded-lg border transition-all
                  ${lineHeight === h
                    ? 'border-blue-500 bg-blue-50 text-blue-700 shadow-sm'
                    : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50 text-gray-500'
                  }`}
              >
                {h === 1.6 ? 'Compact' : h === 1.8 ? 'Normal' : 'Relaxed'}
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
