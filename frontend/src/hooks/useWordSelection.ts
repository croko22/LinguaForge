import { useState, useEffect, useCallback, type MouseEvent } from "react";
import { loadWords, saveWord } from "../api/words";

interface UseWordSelectionOptions {
  documentId: string | undefined;
  sourceLang: string;
  targetLang: string;
  onWordSelected?: () => void;
}

export function useWordSelection({
  documentId,
  sourceLang,
  targetLang,
  onWordSelected,
}: UseWordSelectionOptions) {
  const [selectedWord, setSelectedWord] = useState<string | null>(null);
  const [popoverPos, setPopoverPos] = useState<{
    x: number;
    y: number;
  } | null>(null);
  const [clickedWords, setClickedWords] = useState<string[]>([]);

  // Load previously saved words on mount
  useEffect(() => {
    loadWords()
      .then((words) => {
        setClickedWords(words.map((w) => w.word));
      })
      .catch(() => {});
  }, []);

  const handleWordClick = useCallback(
    (word: string, e: MouseEvent<HTMLSpanElement>) => {
      const clean = word.replace(/^[^\w]+|[^\w]+$/g, "");
      if (!clean) return;
      const rect = (e.target as HTMLElement).getBoundingClientRect();
      setPopoverPos({
        x: rect.left + rect.width / 2,
        y: rect.bottom + 4,
      });
      setSelectedWord(clean);
      setClickedWords((prev) =>
        prev.includes(clean) ? prev : [...prev, clean],
      );
      onWordSelected?.();
      saveWord({
        word: clean,
        translation: "",
        documentId: documentId ?? "",
        sourceLang,
        targetLang,
      });
    },
    [documentId, sourceLang, targetLang, onWordSelected],
  );

  const clearSelection = useCallback(() => {
    setSelectedWord(null);
    setPopoverPos(null);
  }, []);

  const clearWords = useCallback(() => setClickedWords([]), []);

  return {
    selectedWord,
    popoverPos,
    clickedWords,
    handleWordClick,
    clearSelection,
    clearWords,
  } as const;
}
