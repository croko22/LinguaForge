import { useEffect, type RefObject } from "react";

interface UseKeyboardNavigationOptions {
  maxPage: number;
  setCurrentPage: React.Dispatch<React.SetStateAction<number>>;
  setShowChapters: React.Dispatch<React.SetStateAction<boolean>>;
  setShowSettings: React.Dispatch<React.SetStateAction<boolean>>;
  chapterMenuRef: RefObject<HTMLDivElement | null>;
}

export function useKeyboardNavigation({
  maxPage,
  setCurrentPage,
  setShowChapters,
  setShowSettings,
  chapterMenuRef,
}: UseKeyboardNavigationOptions) {
  // Arrow / Page keys for page navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement).tagName;
      if (tag === "INPUT" || tag === "SELECT" || tag === "TEXTAREA") return;
      if (e.key === "ArrowLeft" || e.key === "PageUp") {
        setCurrentPage((p) => Math.max(0, p - 1));
      } else if (e.key === "ArrowRight" || e.key === "PageDown") {
        setCurrentPage((p) => Math.min(maxPage, p + 1));
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [maxPage, setCurrentPage]);

  // Escape key + click-outside to close popovers / menus
  useEffect(() => {
    const handleClickOutside = (e: globalThis.MouseEvent) => {
      if (
        chapterMenuRef.current &&
        !chapterMenuRef.current.contains(e.target as Node)
      ) {
        setShowChapters(false);
      }
    };
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        setShowChapters(false);
        setShowSettings(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    document.addEventListener("keydown", handleEscape);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleEscape);
    };
  }, [chapterMenuRef, setShowChapters, setShowSettings]);
}
