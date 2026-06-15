import { useState, useEffect, useRef } from "react";

export function useViewportResize() {
  const [viewportHeight, setViewportHeight] = useState(0);
  const [viewportWidth, setViewportWidth] = useState(0);
  const viewportRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const calculate = () => {
      if (!viewportRef.current) return;
      setViewportHeight(viewportRef.current.clientHeight);
      setViewportWidth(viewportRef.current.clientWidth);
    };

    const frame = window.requestAnimationFrame(calculate);

    const observer = new ResizeObserver(calculate);
    if (viewportRef.current) observer.observe(viewportRef.current);
    return () => {
      window.cancelAnimationFrame(frame);
      observer.disconnect();
    };
  }, []);

  return { viewportRef, viewportHeight, viewportWidth } as const;
}
