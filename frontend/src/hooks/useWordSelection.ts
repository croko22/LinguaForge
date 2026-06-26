import { useState, useEffect, useCallback, type MouseEvent } from "react";
import { loadWords, saveWord } from "../api/words";
import { translateWord } from "../api/translate";

export function normalizeToken(token: string): string {
  return token.replace(/^[^\w]+|[^\w]+$/g, "");
}

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
  const [translations, setTranslations] = useState<Record<string, string>>({});
  const [lookup, setLookup] = useState<{
    word: string | null;
    translation: string | null;
    loading: boolean;
    error: boolean;
  }>({ word: null, translation: null, loading: false, error: false });

  // Load previously saved words on mount
  useEffect(() => {
    loadWords()
      .then((words) => {
        setClickedWords(words.map((w) => w.word));
        setTranslations(
          words.reduce<Record<string, string>>((acc, word) => {
            if (word.translation) acc[word.word] = word.translation;
            return acc;
          }, {}),
        );
      })
      .catch(() => {});
  }, []);

  const lookupAndSave = useCallback(
    async (word: string) => {
      let translation = translations[word] ?? "";

      if (!translation) {
        try {
          const result = await translateWord({
            word,
            source_lang: sourceLang,
            target_lang: targetLang,
          });
          translation = result.translation;
          setTranslations((prev) => ({ ...prev, [word]: translation }));
          setLookup((prev) =>
            prev.word === word
              ? { word, translation, loading: false, error: false }
              : prev,
          );
        } catch {
          setLookup((prev) =>
            prev.word === word
              ? { word, translation: null, loading: false, error: true }
              : prev,
          );
        }
      }

      try {
        const saved = await saveWord({
          word,
          translation,
          documentId: documentId ?? "",
          sourceLang,
          targetLang,
        });
        if (saved.translation) {
          setTranslations((prev) => ({ ...prev, [word]: saved.translation }));
          setLookup((prev) =>
            prev.word === word
              ? {
                  word,
                  translation: saved.translation,
                  loading: false,
                  error: false,
                }
              : prev,
          );
        }
      } catch {
        if (!translation) {
          setLookup((prev) =>
            prev.word === word
              ? { word, translation: null, loading: false, error: true }
              : prev,
          );
        }
      }
    },
    [documentId, sourceLang, targetLang, translations],
  );

  const handleWordClick = useCallback(
    (word: string, e: MouseEvent<HTMLSpanElement>) => {
      const clean = normalizeToken(word);
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
      setLookup({
        word: clean,
        translation: translations[clean] ?? null,
        loading: !translations[clean],
        error: false,
      });
      onWordSelected?.();
      void lookupAndSave(clean);
    },
    [lookupAndSave, onWordSelected, translations],
  );

  const clearSelection = useCallback(() => {
    setSelectedWord(null);
    setPopoverPos(null);
    setLookup((prev) => ({ ...prev, word: null, loading: false, error: false }));
  }, []);

  const clearWords = useCallback(() => setClickedWords([]), []);

  const isCurrentLookup = lookup.word === selectedWord;

  return {
    selectedWord,
    popoverPos,
    clickedWords,
    translations,
    selectedTranslation: isCurrentLookup
      ? lookup.translation
      : selectedWord
        ? translations[selectedWord] ?? null
        : null,
    selectedTranslationLoading: isCurrentLookup ? lookup.loading : false,
    selectedTranslationError: isCurrentLookup ? lookup.error : false,
    handleWordClick,
    clearSelection,
    clearWords,
  } as const;
}
