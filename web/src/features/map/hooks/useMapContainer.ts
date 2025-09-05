import { useEffect, useRef, useState } from "react";

interface ContainerDimensions {
  width: number;
  height: number;
}

export function useMapContainer() {
  const containerRef = useRef<HTMLDivElement>(null);
  const [dimensions, setDimensions] = useState<ContainerDimensions | null>(null);
  const [isReady, setIsReady] = useState(false);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    let resizeObserver: ResizeObserver | null = null;

    const checkDimensions = () => {
      const rect = container.getBoundingClientRect();
      const width = rect.width;
      const height = rect.height;

      if (width > 0 && height > 0) {
        setDimensions({ width, height });
        setIsReady(true);
        return true;
      }
      return false;
    };

    // Try immediate check
    if (!checkDimensions()) {
      // If not ready, use ResizeObserver to wait for dimensions
      resizeObserver = new ResizeObserver((entries) => {
        for (const entry of entries) {
          const { width, height } = entry.contentRect;
          if (width > 0 && height > 0) {
            setDimensions({ width, height });
            setIsReady(true);
            resizeObserver?.disconnect();
            break;
          }
        }
      });

      resizeObserver.observe(container);
    }

    return () => {
      resizeObserver?.disconnect();
    };
  }, []);

  return {
    containerRef,
    dimensions,
    isReady,
  };
}