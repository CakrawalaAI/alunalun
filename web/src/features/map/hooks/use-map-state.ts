import { create } from "zustand";
import { persist } from "zustand/middleware";
import { MAP_CONFIG } from "@/features/map/constants/map-config";

export interface MapViewState {
  center: [number, number]; // [lng, lat]
  zoom: number;
  bearing: number; // 0-360 degrees
  pitch: number; // 0-60 degrees
}

interface MapStore {
  // View state
  viewState: MapViewState;
  hasLoadedSavedState: boolean;

  // Actions
  setViewState: (state: Partial<MapViewState>) => void;
  loadSavedState: () => MapViewState | null;
  resetToDefault: () => void;
  clearSavedState: () => void;
}

const DEFAULT_VIEW_STATE: MapViewState = {
  center: MAP_CONFIG.initialView.center,
  zoom: MAP_CONFIG.initialView.zoom,
  bearing: MAP_CONFIG.initialView.bearing ?? 0,
  pitch: MAP_CONFIG.initialView.pitch ?? 0,
};

export const useMapState = create<MapStore>()(
  persist(
    (set, get) => ({
      // Initial state
      viewState: DEFAULT_VIEW_STATE,
      hasLoadedSavedState: false,

      // Set view state (partial updates allowed)
      setViewState: (newState) =>
        set((state) => ({
          viewState: { ...state.viewState, ...newState },
        })),

      // Load saved state from localStorage
      loadSavedState: () => {
        const state = get();
        if (!state.hasLoadedSavedState) {
          set({ hasLoadedSavedState: true });
          return state.viewState;
        }
        return null;
      },

      // Reset to default DPR Jakarta view
      resetToDefault: () =>
        set({
          viewState: DEFAULT_VIEW_STATE,
        }),

      // Clear saved state from localStorage
      clearSavedState: () => {
        localStorage.removeItem("alunalun-map-state");
        set({
          viewState: DEFAULT_VIEW_STATE,
          hasLoadedSavedState: false,
        });
      },
    }),
    {
      name: "alunalun-map-state",
      partialize: (state) => ({ viewState: state.viewState }),
    },
  ),
);

// Debounce utility for saving map state
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number,
): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout> | null = null;

  return (...args: Parameters<T>) => {
    if (timeout) {
      clearTimeout(timeout);
    }
    timeout = setTimeout(() => func(...args), wait);
  };
}
