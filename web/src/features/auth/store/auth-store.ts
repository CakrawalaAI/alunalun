import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface User {
  id: string;
  email: string;
  username: string;
  displayName?: string;
  avatarUrl?: string;
  createdAt: string;
  updatedAt: string;
}

type AuthState = {
  type: "none" | "anonymous" | "authenticated";
  token: string | null;
  sessionId: string | null;
  username: string | null;
  user: User | null;

  // Actions
  setAnonymousAuth: (
    token: string,
    sessionId: string,
    username: string,
  ) => void;
  setAuthenticatedAuth: (token: string, user: User) => void;
  clearAuth: () => void;
  getAuthHeaders: () => Record<string, string>;
  isAuthenticated: () => boolean;
  isAnonymous: () => boolean;
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      type: "none",
      token: null,
      sessionId: null,
      username: null,
      user: null,

      setAnonymousAuth: (token, sessionId, username) => {
        set({
          type: "anonymous",
          token,
          sessionId,
          username,
          user: null,
        });
      },

      setAuthenticatedAuth: (token, user) => {
        // Clear anonymous data
        localStorage.removeItem("anonymous_token");
        set({
          type: "authenticated",
          token,
          sessionId: null,
          username: user.username,
          user,
        });
      },

      clearAuth: () => {
        set({
          type: "none",
          token: null,
          sessionId: null,
          username: null,
          user: null,
        });
      },

      getAuthHeaders: () => {
        const state = get();
        if (state.token) {
          return { Authorization: `Bearer ${state.token}` };
        }
        return {} as Record<string, string>;
      },

      isAuthenticated: () => {
        const state = get();
        return state.type === "authenticated";
      },

      isAnonymous: () => {
        const state = get();
        return state.type === "anonymous";
      },
    }),
    {
      name: "auth-storage",
      partialize: (state) => ({
        // Only persist anonymous token and session info
        type: state.type,
        token: state.type === "anonymous" ? state.token : null,
        sessionId: state.type === "anonymous" ? state.sessionId : null,
        username: state.type === "anonymous" ? state.username : null,
      }),
    },
  ),
);
