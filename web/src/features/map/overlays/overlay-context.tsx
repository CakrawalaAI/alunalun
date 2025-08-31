import {
  createContext,
  useContext,
  useState,
  useCallback,
  type ReactNode,
} from "react";
import { OverlayContainer } from "./overlay-container";

interface Overlay {
  id: string;
  component: ReactNode;
  position?: { lat: number; lng: number };
  zIndex?: number;
}

interface OverlayContextValue {
  overlays: Overlay[];
  addOverlay: (component: ReactNode, options?: {
    id?: string;
    position?: { lat: number; lng: number };
    zIndex?: number;
  }) => string;
  removeOverlay: (id: string) => void;
  updateOverlay: (id: string, component: ReactNode) => void;
  clearOverlays: () => void;
}

const OverlayContext = createContext<OverlayContextValue | null>(null);

interface OverlayProviderProps {
  children: ReactNode;
}

/**
 * Provider for managing DOM overlays on the map
 * Handles external React components that need to be rendered on top of the map
 */
export function OverlayProvider({ children }: OverlayProviderProps) {
  const [overlays, setOverlays] = useState<Overlay[]>([]);

  const addOverlay = useCallback((
    component: ReactNode,
    options?: {
      id?: string;
      position?: { lat: number; lng: number };
      zIndex?: number;
    }
  ) => {
    const id = options?.id || `overlay-${Date.now()}-${Math.random()}`;
    const newOverlay: Overlay = {
      id,
      component,
      position: options?.position,
      zIndex: options?.zIndex,
    };
    
    setOverlays(prev => [...prev, newOverlay]);
    return id;
  }, []);

  const removeOverlay = useCallback((id: string) => {
    setOverlays(prev => prev.filter(overlay => overlay.id !== id));
  }, []);

  const updateOverlay = useCallback((id: string, component: ReactNode) => {
    setOverlays(prev =>
      prev.map(overlay =>
        overlay.id === id ? { ...overlay, component } : overlay
      )
    );
  }, []);

  const clearOverlays = useCallback(() => {
    setOverlays([]);
  }, []);

  return (
    <OverlayContext.Provider
      value={{
        overlays,
        addOverlay,
        removeOverlay,
        updateOverlay,
        clearOverlays,
      }}
    >
      {children}
      <OverlayContainer>
        {overlays
          .sort((a, b) => (a.zIndex || 0) - (b.zIndex || 0))
          .map(overlay => (
            <div
              key={overlay.id}
              style={{
                position: overlay.position ? "absolute" : "relative",
                zIndex: overlay.zIndex,
              }}
            >
              {overlay.component}
            </div>
          ))}
      </OverlayContainer>
    </OverlayContext.Provider>
  );
}

/**
 * Hook to access overlay management functions
 * Use this to add/remove/update DOM overlays on the map
 */
export function useMapOverlays() {
  const context = useContext(OverlayContext);
  if (!context) {
    throw new Error("useMapOverlays must be used within OverlayProvider");
  }
  return context;
}