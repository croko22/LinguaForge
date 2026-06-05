import { Routes, Route, Link, useLocation } from 'react-router-dom'
import LibraryPage from './pages/LibraryPage'
import ReaderPage from './pages/ReaderPage'
import SettingsPage from './pages/SettingsPage'
import ReviewPage from './pages/ReviewPage'
import { useDueCount } from './hooks/useReview'

function NavBadge() {
  const { data: count } = useDueCount()
  if (!count || count === 0) return null
  return (
    <span className="inline-flex items-center justify-center w-5 h-5 text-[11px] font-bold bg-surface text-primary rounded-full">
      {count > 99 ? '99+' : count}
    </span>
  )
}

function App() {
  const { pathname } = useLocation()

  return (
    <div className="min-h-screen bg-surface-muted">
      <nav className="sticky top-0 z-30 backdrop-blur-md bg-surface-glass border-b border-border px-6 py-3 flex items-center justify-between">
        <Link
          to="/"
          className={`font-bold text-lg transition-colors ${pathname === '/' ? 'text-primary' : 'text-text hover:text-primary'}`}
        >
          LinguaForge
        </Link>
        <div className="flex items-center gap-3">
          <Link
            to="/review"
            className={`flex items-center gap-1.5 px-4 py-1.5 text-sm font-medium rounded-full transition-colors ${
               pathname === '/review'
                 ? 'bg-primary-hover text-text-inverse'
                 : 'bg-primary text-text-inverse hover:bg-primary-hover'
            }`}
          >
            Review
            <NavBadge />
          </Link>
          <Link
            to="/settings"
            className={`p-2 rounded-full transition-colors ${
               pathname === '/settings'
                 ? 'text-primary bg-primary-light'
                 : 'text-text-muted hover:text-text-secondary hover:bg-surface-hover'
            }`}
            title="Settings"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
          </Link>
        </div>
      </nav>
      <Routes>
        <Route path="/" element={<LibraryPage />} />
        <Route path="/read/:id" element={<ReaderPage />} />
        <Route path="/read/:id/:chapterIndex" element={<ReaderPage />} />
        <Route path="/review" element={<ReviewPage />} />
        <Route path="/settings" element={<SettingsPage />} />
      </Routes>
    </div>
  )
}

export default App
