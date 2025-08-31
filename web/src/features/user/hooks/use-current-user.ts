import { useQuery } from "@tanstack/react-query";
import { userClient } from "@/common/services/connectrpc/client";
import { protoTimestampToISO } from "@/common/utils/timestamp";
import { useAuthStore } from "@/features/auth/store/auth-store";
import type { User } from "@/features/auth/types";

export const useCurrentUser = () => {
  const { type, token } = useAuthStore((state) => ({
    type: state.type,
    token: state.token,
  }));

  return useQuery({
    queryKey: ["currentUser"],
    queryFn: async (): Promise<User | null> => {
      // Only fetch if authenticated (not anonymous)
      if (type !== "authenticated" || !token) {
        return null;
      }

      try {
        const response = await userClient.getCurrentUser({});

        if (!response.user) {
          return null;
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
      } catch (error) {
        console.error("Failed to fetch current user:", error);
        return null;
      }
    },
    enabled: type === "authenticated" && !!token,
    staleTime: 5 * 60 * 1000, // Consider data stale after 5 minutes
    gcTime: 10 * 60 * 1000, // Keep in cache for 10 minutes (formerly cacheTime)
  });
};
