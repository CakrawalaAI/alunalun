import { useMutation } from "@tanstack/react-query";
import { authClient } from "@/common/services/connectrpc/client";
import { protoTimestampToISO } from "@/common/utils/timestamp";
import { useAuthStore } from "@/features/auth/store/auth-store";
import type { AuthenticateResponse, User } from "@/features/auth/types";

interface AuthenticateParams {
  provider: string;
  credential: string;
}

export const useAuthenticate = () => {
  const { sessionId, setAuthenticatedAuth } = useAuthStore((state) => ({
    sessionId: state.sessionId,
    setAuthenticatedAuth: state.setAuthenticatedAuth,
  }));

  return useMutation({
    mutationFn: async ({
      provider,
      credential,
    }: AuthenticateParams): Promise<AuthenticateResponse> => {
      const response = await authClient.authenticate({
        provider,
        credential,
        sessionId: sessionId || undefined, // Include session ID for migration if available
      });

      // Map the protobuf user to our User type
      const user: User = {
        id: response.user?.id || "",
        email: response.user?.email || "",
        username: response.user?.username || "",
        displayName: response.user?.displayName,
        avatarUrl: response.user?.avatarUrl,
        createdAt: protoTimestampToISO(response.user?.createdAt),
        updatedAt: protoTimestampToISO(response.user?.updatedAt),
      };

      return {
        token: response.token,
        user,
        sessionMigrated: response.sessionMigrated,
      };
    },
    onSuccess: (data) => {
      // Store the authenticated auth in the store
      setAuthenticatedAuth(data.token, data.user);

      // Show a success message if session was migrated
      if (data.sessionMigrated) {
        // You can add a toast notification here
        console.log(
          "Your anonymous content has been successfully migrated to your account!",
        );
      }
    },
    onError: (error) => {
      console.error("Authentication failed:", error);
    },
  });
};
