const API_BASE = '/api'

export interface ProviderConfig {
  name: string
  endpoint: string
  api_key?: string
}

export interface Settings {
  active_provider: string
  providers: ProviderConfig[]
}

export async function fetchSettings(): Promise<Settings> {
  const res = await fetch(`${API_BASE}/settings`)
  if (!res.ok) throw new Error('Failed to fetch settings')
  return res.json()
}

export async function updateSettings(settings: Settings): Promise<Settings> {
  const res = await fetch(`${API_BASE}/settings`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(settings),
  })
  if (!res.ok) throw new Error('Failed to update settings')
  return res.json()
}
