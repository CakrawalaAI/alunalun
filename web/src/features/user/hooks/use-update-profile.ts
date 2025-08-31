import { useMutation, useQueryClient } from "@tanstack/react-query";
import { userClient } from "@/common/services/connectrpc/client";
import { protoTimestampToISO } from "@/common/utils/timestamp";
import { useAuthStore } from "@/features/auth/store/auth-store";
import type { User } from "@/features/auth/types";

interface UpdateProfileParams {
  displayName?: string;
  avatarUrl?: string;
}

export const useUpdateProfile = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (params: UpdateProfileParams): Promise<User> => {
      const response = await userClient.updateProfile({
        displayName: params.displayName,
        avatarUrl: params.avatarUrl,
      });

      if (!response.user) {
        throw new Error("Failed to update profile");
      }

      return {
        id: response.user.id,
        email: response.user.email,
        username: response.user.username,
        displayName: response.user.displayName,
        avatarUrl: response.user.avatarUrl,
        createdAt: protoTimestampToISO(response.user.createdAt),
        updatedAt: protoTimestampToISO(response.user.updatedAt),
      };
    },
    onSuccess: (updatedUser) => {
      // Update the user in the auth store
      const token = useAuthStore.getState().token;
      if (token) {
        useAuthStore.getState().setAuthenticatedAuth(token, updatedUser);
      }

      // Invalidate and refetch current user query
      queryClient.invalidateQueries({ queryKey: ["currentUser"] });
    },
    onError: (error) => {
      console.error("Failed to update profile:", error);
    },
  });
};
