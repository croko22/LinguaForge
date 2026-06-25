import { useState, useEffect } from 'react'
import { fetchSettings, updateSettings, type Settings } from '../api/settings'

export default function SettingsPage() {
  const [settings, setSettings] = useState<Settings | null>(null)
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState('')

  useEffect(() => {
    fetchSettings().then(setSettings).catch(() => setMessage('Failed to load settings'))
  }, [])

  const handleProviderChange = (name: string) => {
    if (!settings) return
    setSettings({ ...settings, active_provider: name })
  }

  const handleEndpointChange = (name: string, endpoint: string) => {
    if (!settings) return
    setSettings({
      ...settings,
      providers: settings.providers.map((p) => (p.name === name ? { ...p, endpoint } : p)),
    })
  }

  const handleApiKeyChange = (name: string, apiKey: string) => {
    if (!settings) return
    setSettings({
      ...settings,
      providers: settings.providers.map((p) => (p.name === name ? { ...p, api_key: apiKey } : p)),
    })
  }

  const handleSave = async () => {
    if (!settings) return
    setSaving(true)
    setMessage('')
    try {
      await updateSettings(settings)
      setMessage('Settings saved')
    } catch {
      setMessage('Failed to save settings')
    } finally {
      setSaving(false)
    }
  }

  if (!settings) {
    return (
      <div className="max-w-2xl mx-auto p-6">
        <p className="text-text-secondary">Loading settings...</p>
      </div>
    )
  }

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-6">Settings</h1>

      {/* Active provider */}
      <div className="mb-6">
        <label className="block text-sm font-medium mb-2">Translation Provider</label>
        <div className="flex gap-3">
          {settings.providers.map((p) => (
            <button
              key={p.name}
              onClick={() => handleProviderChange(p.name)}
              className={`px-4 py-2 rounded border text-sm capitalize ${
                settings.active_provider === p.name
                   ? 'bg-primary text-text-inverse border-primary'
                   : 'hover:bg-surface-hover'
              }`}
            >
              {p.name}
            </button>
          ))}
        </div>
      </div>

      {/* Provider configs */}
      {settings.providers.map((p) => (
        <div
          key={p.name}
          className={`mb-4 p-4 border rounded ${
            settings.active_provider === p.name
               ? 'border-primary bg-primary-light'
               : 'opacity-60'
          }`}
        >
          <h3 className="font-semibold capitalize mb-3">{p.name}</h3>

          {p.name === 'deepl' && (
            <>
              <div className="mb-3">
                <label className="block text-xs text-text-secondary mb-1">API URL</label>
                <input
                  type="text"
                  value={p.endpoint}
                  onChange={(e) => handleEndpointChange(p.name, e.target.value)}
                  className="w-full border rounded px-3 py-2 text-sm"
                  placeholder="https://api-free.deepl.com"
                />
              </div>
              <div className="mb-3">
                <label className="block text-xs text-text-secondary mb-1">Authentication Key</label>
                <input
                  type="password"
                  value={p.api_key ?? ''}
                  onChange={(e) => handleApiKeyChange(p.name, e.target.value)}
                  className="w-full border rounded px-3 py-2 text-sm"
                  placeholder="DeepL-Auth-Key xxxx-xxxx-xxxx-xxxx"
                />
                <p className="text-xs text-text-muted mt-1">
                  Get your key at{' '}
                  <a href="https://www.deepl.com/pro-api" className="underline" target="_blank" rel="noreferrer">
                    deepl.com/pro-api
                  </a>
                </p>
              </div>
            </>
          )}

          {p.name === 'libre' && (
            <>
              <div className="mb-3">
                <label className="block text-xs text-text-secondary mb-1">API Endpoint</label>
                <input
                  type="text"
                  value={p.endpoint}
                  onChange={(e) => handleEndpointChange(p.name, e.target.value)}
                  className="w-full border rounded px-3 py-2 text-sm"
                  placeholder="https://libretranslate.com"
                />
              </div>
              <div className="mb-3">
                <label className="block text-xs text-text-secondary mb-1">API Key (optional)</label>
                <input
                  type="password"
                  value={p.api_key ?? ''}
                  onChange={(e) => handleApiKeyChange(p.name, e.target.value)}
                  className="w-full border rounded px-3 py-2 text-sm"
                  placeholder="Leave empty for public instance"
                />
              </div>
            </>
          )}

          {p.name === 'mock' && (
            <p className="text-sm text-text-secondary">
              Embedded dictionary (works offline, 24 words en→es). No configuration needed.
            </p>
          )}
        </div>
      ))}

      {/* Save */}
      <button
        onClick={handleSave}
        disabled={saving}
        className="px-6 py-2 bg-primary text-text-inverse rounded hover:bg-primary-hover disabled:opacity-50"
      >
        {saving ? 'Saving...' : 'Save'}
      </button>

      {message && (
        <p
          className={`mt-3 text-sm ${message.includes('Failed') ? 'text-danger' : 'text-green-600'}`}
        >
          {message}
        </p>
      )}
    </div>
  )
}
