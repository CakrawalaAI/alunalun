import { useCallback } from "react";
import { useAuthStore } from "@/features/auth/store/auth-store";

export const useAuth = () => {
  const {
    type,
    token,
    sessionId,
    username,
    user,
    clearAuth,
    isAuthenticated,
    isAnonymous,
  } = useAuthStore();

  // Check if user has any auth (anonymous or authenticated)
  const hasAuth = useCallback(() => {
    return type !== "none" && token !== null;
  }, [type, token]);

  // Get display name for the current user
  const getDisplayName = useCallback(() => {
    if (user?.displayName) {
      return user.displayName;
    }
    if (username) {
      return username;
    }
    return "Anonymous";
  }, [user, username]);

  // Sign out function
  const signOut = useCallback(() => {
    clearAuth();
    // Optionally redirect to home or refresh the page
    // window.location.href = '/';
  }, [clearAuth]);

  return {
    // Auth state
    type,
    token,
    sessionId,
    username,
    user,

    // Auth checks
    hasAuth,
    isAuthenticated,
    isAnonymous,

    // Utility functions
    getDisplayName,
    signOut,
  };
};
