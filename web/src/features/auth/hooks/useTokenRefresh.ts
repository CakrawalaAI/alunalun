import { useMutation } from "@tanstack/react-query";
import { useAuthStore } from "../store/authStore";
import { refreshToken } from "@/common/services/connectrpc";
import { create } from "@bufbuild/protobuf";
import { RefreshTokenRequestSchema } from "@/common/services/connectrpc/v1/service/auth_pb";

export function useTokenRefresh() {
  const { token, setAuth, user } = useAuthStore();

  const refreshTokenMutation = useMutation({
    mutationFn: async (expiredToken?: string) => {
      const tokenToRefresh = expiredToken || token;
      
      if (!tokenToRefresh) {
        throw new Error("No token to refresh");
      }

      const request = create(RefreshTokenRequestSchema, {
        expiredToken: tokenToRefresh,
      });

      const response = await refreshToken(request);
      return response;
    },
    onSuccess: (response) => {
      // Update the token in the auth store while keeping the same user
      if (user) {
        setAuth(user, response.token);
      }
    },
    onError: (error) => {
      console.error("Token refresh failed:", error);
      // If refresh fails, the user needs to re-authenticate
      throw error;
    },
  });

  const refresh = async (expiredToken?: string) => {
    await refreshTokenMutation.mutateAsync(expiredToken);
  };

  return {
    refresh,
    isLoading: refreshTokenMutation.isPending,
    error: refreshTokenMutation.error,
  };
}