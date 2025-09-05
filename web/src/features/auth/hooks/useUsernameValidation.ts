import { useState, useCallback } from "react";
import { useMutation } from "@tanstack/react-query";
import { checkUsername } from "@/common/services/connectrpc";
import { create } from "@bufbuild/protobuf";
import { CheckUsernameRequestSchema } from "@/common/services/connectrpc/v1/service/auth_pb";

export function useUsernameValidation() {
  const [isAvailable, setIsAvailable] = useState<boolean | null>(null);

  const checkUsernameMutation = useMutation({
    mutationFn: async (username: string) => {
      const request = create(CheckUsernameRequestSchema, {
        username: username,
      });

      const response = await checkUsername(request);
      return response;
    },
    onSuccess: (response) => {
      setIsAvailable(response.available);
    },
    onError: (error) => {
      console.error("Username validation failed:", error);
      setIsAvailable(false);
    },
  });

  const validate = useCallback(async (username: string) => {
    setIsAvailable(null);
    await checkUsernameMutation.mutateAsync(username);
  }, [checkUsernameMutation]);

  return {
    isValidating: checkUsernameMutation.isPending,
    isAvailable,
    validate,
  };
}