import { create } from "zustand";

interface UserMenuState {
  isOpen: boolean;
  openMenu: () => void;
  closeMenu: () => void;
  toggle: () => void;
}

export const useUserMenu = create<UserMenuState>((set) => ({
  isOpen: false,
  openMenu: () => set({ isOpen: true }),
  closeMenu: () => set({ isOpen: false }),
  toggle: () => set((state) => ({ isOpen: !state.isOpen })),
}));