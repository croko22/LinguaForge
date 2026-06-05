import { Routes, Route, Link } from 'react-router-dom'
import LibraryPage from './pages/LibraryPage'
import ReaderPage from './pages/ReaderPage'
import SettingsPage from './pages/SettingsPage'
import ReviewPage from './pages/ReviewPage'
import { useDueCount } from './hooks/useReview'

function NavBadge() {
  const { data: count } = useDueCount()
  if (!count || count === 0) return null
  return (
    <span className="ml-1.5 inline-flex items-center justify-center w-5 h-5 text-[11px] font-bold text-white bg-red-500 rounded-full">
      {count > 99 ? '99+' : count}
    </span>
  )
}

function App() {
  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white border-b px-6 py-3 flex items-center justify-between">
        <Link to="/" className="font-bold text-lg">LinguaForge</Link>
        <div className="flex items-center gap-4">
          <Link to="/review" className="text-sm text-gray-500 hover:text-gray-700 flex items-center">
            Review
            <NavBadge />
          </Link>
          <Link to="/settings" className="text-sm text-gray-500 hover:text-gray-700">Settings</Link>
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
