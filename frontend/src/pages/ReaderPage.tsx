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
          <h2 className="text-xl font-semibold text-red-600 mb-2">Document not found</h2>
          <p className="text-gray-500 mb-4">This document may have been deleted or is unavailable.</p>
          <a href="/" className="text-blue-600 hover:underline">← Back to Library</a>
        </div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-center">
          <h2 className="text-xl font-semibold text-red-600 mb-2">Failed to load chapter</h2>
          <p className="text-gray-500 mb-4">There was a problem loading this chapter.</p>
          <a href="/" className="text-blue-600 hover:underline">← Back to Library</a>
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
        <p className="text-gray-500">Loading...</p>
      </div>
    );
  }

  const themeClass = theme === 'light' ? 'reader-light' : theme === 'sepia' ? 'reader-sepia' : 'reader-dark';

  return (
    <div className="flex h-screen">
      {/* Text area */}
      <div className="flex-1 flex flex-col overflow-hidden relative">
        {/* Chapter header */}
        <div className="border-b p-4 flex items-center justify-between">
          <h1 className="text-xl font-semibold">{chapter.chapter_title}</h1>
          <div className="flex items-center gap-2">
            <button
              onClick={() => goTo(currentChapter - 1)}
              disabled={!hasPrev}
              className="px-3 py-1 border rounded disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
            >
              Prev
            </button>
            <select
              value={currentChapter}
              onChange={(e) => goTo(parseInt(e.target.value, 10))}
              className="border rounded px-2 py-1"
            >
              {chapters?.map((ch) => (
                <option key={ch.chapter_index} value={ch.chapter_index}>
                  {ch.chapter_title} content
                </option>
              ))}
            </select>
            <button
              onClick={() => goTo(currentChapter + 1)}
              disabled={!hasNext}
              className="px-3 py-1 border rounded disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
            >
              Next
            </button>
            <button
              onClick={() => setShowSettings(s => !s)}
              className="px-2 py-1 border rounded hover:bg-gray-50 text-sm"
              title="Reader settings"
            >⚙️</button>
          </div>
        </div>

        {/* Chapter content */}
        <div
          className={`flex-1 overflow-y-auto p-6 ${themeClass}`}
          style={{ backgroundColor: 'var(--reader-bg)', color: 'var(--reader-text)' }}
        >
          {showSettings && <SettingsPanel />}
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

      {/* Word panel */}
      <div className="w-80 border-l p-4 overflow-y-auto">
        <WordPanel words={clickedWords} onClear={() => setClickedWords([])} />
      </div>
    </div>
  );
}
