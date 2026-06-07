import { useState, useEffect, useCallback, useMemo, useRef, type MouseEvent } from "react";
import { useParams } from "react-router-dom";
import { useChapters, useChapterContent, useDocument } from "../hooks/useReader";
import { useReaderSettings } from "../store/readerSettings";
import { getReadingProgress, getReadingProgressFromDB, setReadingProgress } from "../hooks/useReadingProgress";
import TextDisplay from "../components/TextDisplay";
import WordPopover from "../components/WordPopover";
import WordPanel from "../components/WordPanel";
import SettingsPanel from "../components/SettingsPanel";
import { loadWords, saveWord } from "../api/words";
import { paginateReaderContent } from "../services/readerPagination";

export default function ReaderPage() {
  const { id, chapterIndex: chapterIndexParam } = useParams();
  const [currentChapter, setCurrentChapter] = useState(() => {
    if (chapterIndexParam) return parseInt(chapterIndexParam, 10)
    if (id) {
      const saved = getReadingProgress(id)
      if (saved !== null) return saved
    }
    return 0
  });

  useEffect(() => {
    if (!id || chapterIndexParam || initialLoadDone.current) return
    initialLoadDone.current = true
    getReadingProgressFromDB(id).then(saved => {
      if (saved !== null) {
        setCurrentChapter(saved)
      }
    }).catch(() => {})
  }, [id, chapterIndexParam])
  const [showSettings, setShowSettings] = useState(false);
  const [showChapters, setShowChapters] = useState(false);
  const [showSidebar, setShowSidebar] = useState(false);
  const { theme, fontSize, lineHeight } = useReaderSettings();

  const { data: book, isError: documentError } = useDocument(id ?? "");
  const { data: chapters, isError: chaptersError } = useChapters(id ?? "");
  const { data: chapter, isError } = useChapterContent(id ?? "", currentChapter);

  const [selectedWord, setSelectedWord] = useState<string | null>(null);
  const [popoverPos, setPopoverPos] = useState<{ x: number; y: number } | null>(
    null,
  );
  const [clickedWords, setClickedWords] = useState<string[]>([]);
  const [currentPage, setCurrentPage] = useState(0);
  const [viewportHeight, setViewportHeight] = useState(0);
  const [viewportWidth, setViewportWidth] = useState(0);
  const [isMobile, setIsMobile] = useState(false);
  const viewportRef = useRef<HTMLDivElement>(null);
  const chapterMenuRef = useRef<HTMLDivElement>(null);
  const initialLoadDone = useRef(false);

  const pages = useMemo(
    () =>
      paginateReaderContent(
        chapter?.content ?? '',
        {
          viewportHeight,
          viewportWidth,
          fontSize,
          lineHeight,
          chromeHeight: 96,
        },
      ),
    [chapter?.content, fontSize, lineHeight, viewportHeight, viewportWidth],
  );

  useEffect(() => {
    loadWords().then(words => {
      setClickedWords(words.map(w => w.word))
    }).catch(() => {})
  }, [])

  useEffect(() => {
    const calculate = () => {
      if (!viewportRef.current) return
      setViewportHeight(viewportRef.current.clientHeight)
      setViewportWidth(viewportRef.current.clientWidth)
    }

    const frame = window.requestAnimationFrame(calculate)

    const observer = new ResizeObserver(calculate)
    if (viewportRef.current) observer.observe(viewportRef.current)
    return () => {
      window.cancelAnimationFrame(frame)
      observer.disconnect()
    }
  }, [])

  useEffect(() => {
    const handleClickOutside = (e: globalThis.MouseEvent) => {
      if (chapterMenuRef.current && !chapterMenuRef.current.contains(e.target as Node)) {
        setShowChapters(false)
      }
    }
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setShowChapters(false)
        setShowSettings(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    document.addEventListener('keydown', handleEscape)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
      document.removeEventListener('keydown', handleEscape)
    }
  }, [])

  useEffect(() => {
    if (!window.matchMedia) return
    const mq = window.matchMedia('(max-width: 767px)')
    setIsMobile(mq.matches)
    const handler = (e: MediaQueryListEvent) => setIsMobile(e.matches)
    mq.addEventListener('change', handler)
    return () => mq.removeEventListener('change', handler)
  }, [])

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement).tagName
      if (tag === 'INPUT' || tag === 'SELECT' || tag === 'TEXTAREA') return
      if (e.key === 'ArrowLeft' || e.key === 'PageUp') {
        setCurrentPage((page) => Math.max(0, page - 1))
      } else if (e.key === 'ArrowRight' || e.key === 'PageDown') {
        setCurrentPage((page) => Math.min(Math.max(0, pages.length - 1), page + 1))
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [pages.length])

  const goTo = useCallback((index: number) => {
    setSelectedWord(null);
    setPopoverPos(null);
    setCurrentChapter(index);
    setCurrentPage(0);
    setShowChapters(false)
    if (id) setReadingProgress(id, index);
  }, [id]);

  const totalChapters = chapters?.length ?? 0;
  const hasPrev = currentChapter > 0;
  const hasNext = currentChapter < totalChapters - 1;
  const docTitle = book?.title ?? chapter?.chapter_title ?? 'Reader';
  const totalPages = pages.length;
  const safeCurrentPage = Math.min(currentPage, Math.max(0, totalPages - 1));

  if (chaptersError && !chapters) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <h2 className="text-xl font-semibold text-danger mb-2">Document not found</h2>
          <p className="text-text-secondary mb-4">This document may have been deleted or is unavailable.</p>
          <a href="/" className="text-primary hover:underline">← Back to Library</a>
        </div>
      </div>
    );
  }

  if (documentError && !book) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <h2 className="text-xl font-semibold text-danger mb-2">Failed to load document</h2>
          <p className="text-text-secondary mb-4">There was a problem loading this book.</p>
          <a href="/" className="text-primary hover:underline">← Back to Library</a>
        </div>
      </div>
    )
  }

  if (isError) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <h2 className="text-xl font-semibold text-danger mb-2">Failed to load chapter</h2>
          <p className="text-text-secondary mb-4">There was a problem loading this chapter.</p>
          <a href="/" className="text-primary hover:underline">← Back to Library</a>
        </div>
      </div>
    );
  }

  const handleWordClick = (word: string, e: MouseEvent<HTMLSpanElement>) => {
    const clean = word.replace(/^[^\w]+|[^\w]+$/g, "");
    if (!clean) return;
    const rect = (e.target as HTMLElement).getBoundingClientRect();
    setPopoverPos({ x: rect.left + rect.width / 2, y: rect.bottom + 4 });
    setSelectedWord(clean);
    setClickedWords((prev) => (prev.includes(clean) ? prev : [...prev, clean]));
    setShowSidebar(true);

    saveWord(clean, "", id ?? "");
  };

  const handlePageTurnClick = (e: MouseEvent<HTMLDivElement>) => {
    if ((e.target as HTMLElement).closest('[data-word]')) return
    if (!viewportRef.current) return
    const rect = viewportRef.current.getBoundingClientRect()
    const x = e.clientX - rect.left
    const width = rect.width

    if (x < width * 0.3) {
      setCurrentPage(p => Math.max(0, p - 1))
    } else if (x > width * 0.7) {
      setCurrentPage(p => Math.min(totalPages - 1, p + 1))
    }
  };

  if (!chapter) {
    return (
      <div className="flex h-screen items-center justify-center">
        <p className="text-text-secondary">Loading...</p>
      </div>
    );
  }

  const themeClass = theme === 'light' ? 'reader-light' : theme === 'sepia' ? 'reader-sepia' : 'reader-dark';

  return (
    <div className="flex h-full min-h-0 bg-surface">
      {/* Main area */}
      <div className="flex-1 flex flex-col overflow-hidden relative min-w-0">
        {/* Header - neutral bg, never changes with theme */}
        <header className="bg-surface border-b px-4 py-2 md:py-3 flex items-center justify-between gap-3 shrink-0 relative">
          <div className="flex items-center gap-3 min-w-0 flex-1">
            <a
              href="/"
              className="text-text-muted hover:text-text-secondary transition-colors shrink-0"
              title="Back to library"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
            </a>
            <div className="min-w-0">
              <div className="text-[11px] uppercase tracking-[0.24em] text-text-muted truncate">{docTitle}</div>
              <h1 className="text-lg font-semibold truncate text-text leading-tight">{chapter.chapter_title}</h1>
            </div>
          </div>

          <div className="flex items-center gap-2 shrink-0">
            <button
              onClick={() => setShowSidebar(s => !s)}
              className="px-3 py-1.5 border border-border rounded-lg text-sm text-text-secondary hover:bg-surface-hover transition-colors"
            >
              <span className="md:hidden">{showSidebar ? 'Hide' : 'Show'}</span>
              <span className="hidden md:inline">{showSidebar ? 'Hide vocab' : 'Show vocab'}</span>
            </button>
            <div className="relative" ref={chapterMenuRef}>
              <button
                onClick={() => setShowChapters(s => !s)}
                className="px-3 py-1.5 border border-border rounded-lg text-sm text-text hover:bg-surface-hover transition-colors flex items-center gap-1.5"
                aria-label="Chapters"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                </svg>
                <span className="hidden md:inline">Chapters</span>
              </button>
              {showChapters && (
                <div className="absolute right-0 top-full mt-2 w-80 rounded-2xl border border-border bg-surface shadow-xl p-3 z-50">
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => goTo(currentChapter - 1)}
                      disabled={!hasPrev}
                      className="flex-1 px-3 py-2 border border-border rounded-xl text-sm text-text-secondary disabled:opacity-40 disabled:cursor-not-allowed hover:bg-surface-hover transition-colors"
                    >
                      ← Prev
                    </button>
                    <button
                      onClick={() => goTo(currentChapter + 1)}
                      disabled={!hasNext}
                      className="flex-1 px-3 py-2 border border-border rounded-xl text-sm text-text-secondary disabled:opacity-40 disabled:cursor-not-allowed hover:bg-surface-hover transition-colors"
                    >
                      Next →
                    </button>
                  </div>
                  <div className="mt-3">
                    <label htmlFor="chapter-jump" className="block text-[11px] uppercase tracking-[0.22em] text-text-muted mb-2">Jump to chapter</label>
                    <select
                      id="chapter-jump"
                      value={currentChapter}
                      onChange={(e) => goTo(parseInt(e.target.value, 10))}
                      className="w-full border border-border rounded-xl px-3 py-2 text-sm text-text bg-surface transition-colors"
                    >
                      {chapters?.map((ch) => (
                        <option key={ch.chapter_index} value={ch.chapter_index}>
                          {ch.chapter_title}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>
              )}
            </div>
            <div className="relative">
              <button
                onClick={() => setShowSettings(s => !s)}
                className="w-9 h-9 flex items-center justify-center border border-border rounded-lg hover:bg-surface-hover transition-all text-text-muted"
                title="Reader settings"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
              </button>
              {showSettings && <SettingsPanel onClose={() => setShowSettings(false)} />}
            </div>
          </div>
        </header>

        {/* Content - themed canvas */}
        <div
          ref={viewportRef}
          className={`flex-1 overflow-hidden px-4 sm:px-8 md:px-16 py-8 ${themeClass} theme-transition relative`}
          style={{ backgroundColor: 'var(--reader-bg)', color: 'var(--reader-text)' }}
          onClick={handlePageTurnClick}
        >
          <div
            className="mx-auto w-full max-w-[68ch] relative pb-12 transition-opacity duration-200 ease-out"
            style={{
              fontSize: `${fontSize}rem`,
              lineHeight: lineHeight,
            }}
          >
            <TextDisplay content={pages[safeCurrentPage] ?? chapter.content} onWordClick={handleWordClick} />
          </div>
          <div className="pointer-events-none absolute bottom-3 right-3">
            <div className="rounded-full border border-border/80 bg-surface/95 px-2.5 py-1 text-xs md:text-[10px] font-medium text-text shadow-sm backdrop-blur">
              Page {safeCurrentPage + 1} / {Math.max(1, totalPages)}
            </div>
          </div>
        </div>

        {/* Word popover */}
        {selectedWord && popoverPos && (
          <WordPopover
            word={selectedWord}
            position={popoverPos}
            onClose={() => {
              setSelectedWord(null);
              setPopoverPos(null);
            }}
          />
        )}
      </div>

      {/* Sidebar - Desktop: side-by-side. Mobile: overlay */}
      {showSidebar && !isMobile && (
        <aside className="w-80 border-l bg-surface-muted flex flex-col overflow-y-auto">
          <div className="p-4">
            <WordPanel words={clickedWords} onClear={() => setClickedWords([])} />
          </div>
        </aside>
      )}

      {/* Mobile sidebar overlay with backdrop */}
      {showSidebar && isMobile && (
        <div className="fixed inset-0 z-50">
          <div
            className="fixed inset-0 bg-black/40 backdrop-blur-sm"
            onClick={() => setShowSidebar(false)}
          />
          <div className="fixed right-0 top-0 h-full w-full max-w-sm bg-surface-muted shadow-xl">
            <div className="p-4 overflow-y-auto h-full">
              <WordPanel words={clickedWords} onClear={() => setClickedWords([])} />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
