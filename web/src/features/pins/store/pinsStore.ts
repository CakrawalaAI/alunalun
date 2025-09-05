import { create } from "zustand";

export interface Pin {
  id: string;
  content: string;
  location: {
    latitude: number;
    longitude: number;
  };
  authorId: string;
  authorUsername: string;
  createdAt: string;
  commentCount: number;
  isPending?: boolean; // For optimistic updates
}

interface PinsState {
  // State
  pins: Pin[];
  selectedPin: Pin | null;
  isCreating: boolean;
  creatingLocation: { lat: number; lng: number } | null;
  
  // Actions
  setPins: (pins: Pin[]) => void;
  addPin: (pin: Pin) => void;
  removePin: (id: string) => void;
  selectPin: (pin: Pin | null) => void;
  setCreating: (creating: boolean, location?: { lat: number; lng: number }) => void;
  addOptimisticPin: (pin: Pin) => void;
  removeOptimisticPin: (id: string) => void;
}

export const usePinsStore = create<PinsState>((set) => ({
  pins: [],
  selectedPin: null,
  isCreating: false,
  creatingLocation: null,

  setPins: (pins) => set({ pins }),
  
  addPin: (pin) => set((state) => ({
    pins: [...state.pins, pin],
  })),

  removePin: (id) => set((state) => ({
    pins: state.pins.filter((p) => p.id !== id),
    selectedPin: state.selectedPin?.id === id ? null : state.selectedPin,
  })),

  selectPin: (pin) => set({ selectedPin: pin }),

  setCreating: (creating, location) => set({
    isCreating: creating,
    creatingLocation: location || null,
  }),

  addOptimisticPin: (pin) => set((state) => ({
    pins: [...state.pins, { ...pin, isPending: true }],
  })),

  removeOptimisticPin: (id) => set((state) => ({
    pins: state.pins.filter((p) => p.id !== id),
  })),
}));