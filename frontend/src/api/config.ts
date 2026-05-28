// In Wails desktop mode, use absolute URL to embedded API server
// In Vite dev mode, use relative URL (proxied by Vite)
const isWails = typeof window !== 'undefined' && (window as any).wails !== undefined

export const API_BASE = isWails ? 'http://localhost:8080' : '/api'
