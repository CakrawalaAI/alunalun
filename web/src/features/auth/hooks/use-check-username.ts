import { useMutation } from "@tanstack/react-query";
import { authClient } from "@/common/services/connectrpc/client";
import type { CheckUsernameResponse } from "@/features/auth/types";

export const useCheckUsername = () => {
  return useMutation({
    mutationFn: async (username: string): Promise<CheckUsernameResponse> => {
      const response = await authClient.checkUsername({ username });
      return {
        available: response.available,
        message: response.message,
      };
    },
  });
};