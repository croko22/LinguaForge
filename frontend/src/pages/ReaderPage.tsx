import { useState, useEffect, useCallback, type MouseEvent } from "react";
import { useParams } from "react-router-dom";
import { useChapters, useChapterContent } from "../hooks/useReader";
import { useReaderSettings } from "../store/readerSettings";
import TextDisplay from "../components/TextDisplay";
import WordPopover from "../components/WordPopover";
import WordPanel from "../components/WordPanel";
import SettingsPanel from "../components/SettingsPanel";
import { loadWords, saveWord } from "../api/words";

export default function ReaderPage() {
  const { id, chapterIndex: chapterIndexParam } = useParams();
  const [currentChapter, setCurrentChapter] = useState(() =>
    parseInt(chapterIndexParam ?? "0", 10)
  );
  const [showSettings, setShowSettings] = useState(false);
  const { theme, fontSize, lineHeight } = useReaderSettings();

  const { data: chapters, isError: chaptersError } = useChapters(id ?? "");
  const { data: chapter, isError } = useChapterContent(id ?? "", currentChapter);

  const [selectedWord, setSelectedWord] = useState<string | null>(null);
  const [popoverPos, setPopoverPos] = useState<{ x: number; y: number } | null>(
    null,
  );
  const [clickedWords, setClickedWords] = useState<string[]>([]);

  useEffect(() => {
    loadWords().then(words => {
      setClickedWords(words.map(w => w.word))
    }).catch(() => {})
  }, [])

  const goTo = useCallback((index: number) => {
    setSelectedWord(null);
    setPopoverPos(null);
    setCurrentChapter(index);
  }, []);

  const totalChapters = chapters?.length ?? 0;
  const hasPrev = currentChapter > 0;
  const hasNext = currentChapter < totalChapters - 1;

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

    saveWord(clean, "", id ?? "");
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
    <div className="flex h-screen bg-surface">
      {/* Main area */}
      <div className="flex-1 flex flex-col overflow-hidden relative">
        {/* Header - neutral bg, never changes with theme */}
        <header className="bg-surface border-b px-4 py-2 flex items-center justify-between shrink-0">
          <div className="flex items-center gap-3 min-w-0">
            <a
              href="/"
              className="text-text-muted hover:text-text-secondary transition-colors shrink-0"
              title="Back to library"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
            </a>
            <h1 className="text-base font-semibold truncate text-text">{chapter.chapter_title}</h1>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={() => goTo(currentChapter - 1)}
              disabled={!hasPrev}
              className="px-3 py-1.5 border border-gray-200 rounded-lg text-sm text-gray-600 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-gray-100 hover:border-gray-300 transition-all"
            >
              ← Prev
            </button>
            <select
              value={currentChapter}
              onChange={(e) => goTo(parseInt(e.target.value, 10))}
              className="border border-border rounded-lg px-2 py-1.5 text-sm text-gray-700 bg-surface transition-colors"
            >
              {chapters?.map((ch) => (
                <option key={ch.chapter_index} value={ch.chapter_index}>
                  {ch.chapter_title}
                </option>
              ))}
            </select>
            <button
              onClick={() => goTo(currentChapter + 1)}
              disabled={!hasNext}
              className="px-3 py-1.5 border border-gray-200 rounded-lg text-sm text-gray-600 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-gray-100 hover:border-gray-300 transition-all"
            >
              Next →
            </button>
          </div>

          <div className="relative">
            <button
              onClick={() => setShowSettings(s => !s)}
              className="w-8 h-8 flex items-center justify-center border border-border rounded-lg hover:bg-surface-hover transition-all text-text-muted"
              title="Reader settings"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </button>
            {showSettings && <SettingsPanel onClose={() => setShowSettings(false)} />}
          </div>
        </header>

        {/* Content - themed canvas */}
        <div
          className={`flex-1 overflow-y-auto px-8 py-6 theme-transition ${themeClass}`}
          style={{ backgroundColor: 'var(--reader-bg)', color: 'var(--reader-text)' }}
        >
          <div style={{ fontSize: `${fontSize}rem`, lineHeight: lineHeight }}>
            <TextDisplay
              content={chapter.content}
              onWordClick={handleWordClick}
            />
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

      {/* Sidebar - neutral bg */}
      <aside className="w-80 border-l bg-surface-muted flex flex-col overflow-y-auto">
        <div className="p-4">
          <WordPanel words={clickedWords} onClear={() => setClickedWords([])} />
        </div>
      </aside>
    </div>
  );
}
