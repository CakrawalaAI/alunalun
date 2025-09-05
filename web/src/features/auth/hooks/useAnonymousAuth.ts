import { useMutation } from "@tanstack/react-query";
import { useAuthStore } from "../store/authStore";
import { initAnonymous } from "@/common/services/connectrpc";
import { create } from "@bufbuild/protobuf";
import { InitAnonymousRequestSchema } from "@/common/services/connectrpc/v1/service/auth_pb";

export function useAnonymousAuth() {
  const { setAuth } = useAuthStore();

  const initAnonymousMutation = useMutation({
    mutationFn: async (username: string) => {
      const request = create(InitAnonymousRequestSchema, {
        username: username,
      });

      const response = await initAnonymous(request);
      return response;
    },
    onSuccess: (response) => {
      // Create anonymous user object
      const user = {
        id: response.sessionId, // Use session ID as user ID for anonymous users
        username: response.username,
        email: undefined, // Anonymous users don't have email
        picture: undefined, // Anonymous users don't have profile pictures
        createdAt: new Date().toISOString(),
      };
      
      setAuth(user, response.token);
    },
    onError: (error) => {
      console.error("Anonymous authentication failed:", error);
      throw error;
    },
  });

  const createAnonymousSession = async (username: string) => {
    await initAnonymousMutation.mutateAsync(username);
  };

  return {
    createAnonymousSession,
    isLoading: initAnonymousMutation.isPending,
    error: initAnonymousMutation.error,
  };
}