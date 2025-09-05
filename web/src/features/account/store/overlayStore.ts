import { create } from "zustand";

interface OverlayState {
  isVisible: boolean;
  setVisible: (visible: boolean) => void;
  toggle: () => void;
}

export const useOverlayStore = create<OverlayState>((set) => ({
  isVisible: true,
  setVisible: (visible) => set({ isVisible: visible }),
  toggle: () => set((state) => ({ isVisible: !state.isVisible })),
}));