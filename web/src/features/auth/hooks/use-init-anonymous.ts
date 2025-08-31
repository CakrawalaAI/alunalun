import { useMutation } from "@tanstack/react-query";
import { authClient } from "@/common/services/connectrpc/client";
import { useAuthStore } from "@/features/auth/store/auth-store";
import type { InitAnonymousResponse } from "@/features/auth/types";

export const useInitAnonymous = () => {
  const setAnonymousAuth = useAuthStore((state) => state.setAnonymousAuth);

  return useMutation({
    mutationFn: async (username: string): Promise<InitAnonymousResponse> => {
      const response = await authClient.initAnonymous({ username });
      return {
        token: response.token,
        sessionId: response.sessionId,
        username: response.username,
      };
    },
    onSuccess: (data) => {
      // Store the anonymous auth in the store
      setAnonymousAuth(data.token, data.sessionId, data.username);
    },
    onError: (error) => {
      console.error("Failed to initialize anonymous session:", error);
    },
  });
};
