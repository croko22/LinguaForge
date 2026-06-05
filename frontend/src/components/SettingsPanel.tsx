import { useReaderSettings } from '../store/readerSettings'

export default function SettingsPanel() {
  const { theme, fontSize, lineHeight, setTheme, setFontSize, setLineHeight } = useReaderSettings()

  const fontSteps = [1, 1.125, 1.25, 1.375, 1.5]
  const fontLabels = ['S', 'M', 'L', 'XL', 'XXL']
  const currentFontIndex = fontSteps.indexOf(fontSize)

  return (
    <div className="p-3 border-b space-y-3">
      <div>
        <label className="text-xs text-[var(--reader-muted)] uppercase tracking-wide font-medium">Theme</label>
        <div className="flex gap-2 mt-1">
          {(['light', 'sepia', 'dark'] as const).map((t) => (
            <button
              key={t}
              onClick={() => setTheme(t)}
              className={`flex-1 py-1.5 text-xs font-medium rounded border transition-colors
                ${theme === t
                  ? 'border-blue-500 bg-blue-50 text-blue-700'
                  : 'border-gray-200 hover:bg-gray-50 text-[var(--reader-muted)]'
                }`}
            >
              {t === 'light' ? '☀️ Light' : t === 'sepia' ? '📖 Sepia' : '🌙 Dark'}
            </button>
          ))}
        </div>
      </div>

      <div>
        <label className="text-xs text-[var(--reader-muted)] uppercase tracking-wide font-medium">Font Size</label>
        <div className="flex items-center gap-2 mt-1">
          <button
            onClick={() => currentFontIndex > 0 && setFontSize(fontSteps[currentFontIndex - 1])}
            disabled={currentFontIndex <= 0}
            className="w-8 h-8 flex items-center justify-center border rounded disabled:opacity-30 hover:bg-gray-100"
          >−</button>
          <span className="text-sm font-medium min-w-[2rem] text-center">{fontLabels[currentFontIndex]}</span>
          <button
            onClick={() => currentFontIndex < fontSteps.length - 1 && setFontSize(fontSteps[currentFontIndex + 1])}
            disabled={currentFontIndex >= fontSteps.length - 1}
            className="w-8 h-8 flex items-center justify-center border rounded disabled:opacity-30 hover:bg-gray-100"
          >+</button>
        </div>
      </div>

      <div>
        <label className="text-xs text-[var(--reader-muted)] uppercase tracking-wide font-medium">Spacing</label>
        <div className="flex gap-2 mt-1">
          {([1.6, 1.8, 2.0] as const).map((h) => (
            <button
              key={h}
              onClick={() => setLineHeight(h)}
              className={`flex-1 py-1.5 text-xs font-medium rounded border transition-colors
                ${lineHeight === h
                  ? 'border-blue-500 bg-blue-50 text-blue-700'
                  : 'border-gray-200 hover:bg-gray-50 text-[var(--reader-muted)]'
                }`}
            >
              {h === 1.6 ? 'Compact' : h === 1.8 ? 'Normal' : 'Relaxed'}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
