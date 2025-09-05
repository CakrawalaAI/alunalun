import { create } from "zustand";

interface AuthModalState {
  isOpen: boolean;
  openAuthModal: () => void;
  closeModal: () => void;
  openModal: () => void;
}

export const useAuthModal = create<AuthModalState>((set) => ({
  isOpen: false,
  openAuthModal: () => set({ isOpen: true }),
  openModal: () => set({ isOpen: true }),
  closeModal: () => set({ isOpen: false }),
}));