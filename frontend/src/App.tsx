import { Routes, Route, Link } from 'react-router-dom'
import LibraryPage from './pages/LibraryPage'
import ReaderPage from './pages/ReaderPage'
import SettingsPage from './pages/SettingsPage'

function App() {
  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white border-b px-6 py-3 flex items-center justify-between">
        <Link to="/" className="font-bold text-lg">LinguaForge</Link>
        <Link to="/settings" className="text-sm text-gray-500 hover:text-gray-700">Settings</Link>
      </nav>
      <Routes>
        <Route path="/" element={<LibraryPage />} />
        <Route path="/read/:id" element={<ReaderPage />} />
        <Route path="/read/:id/:chapterIndex" element={<ReaderPage />} />
        <Route path="/settings" element={<SettingsPage />} />
      </Routes>
    </div>
  )
}

export default App
