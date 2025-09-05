import { useMutation, useQueryClient } from "@tanstack/react-query";
import { usePinsStore } from "../store/pinsStore";
import { useAuthStore } from "@/features/auth/store/authStore";
import { createPin } from "@/common/services/connectrpc";
import { create } from "@bufbuild/protobuf";
import { 
  CreatePinRequestSchema,
  type CreatePinRequest 
} from "@/common/services/connectrpc/v1/service/pin_service_pb";
import {
  LocationSchema,
  type Location
} from "@/common/services/connectrpc/v1/entities/pin_pb";

interface CreatePinData {
  content: string;
  location: {
    latitude: number;
    longitude: number;
  };
}

export function useCreatePin() {
  const queryClient = useQueryClient();
  const { addOptimisticPin, removeOptimisticPin, addPin } = usePinsStore();
  const { user } = useAuthStore();

  const mutation = useMutation({
    mutationFn: async (data: CreatePinData) => {
      if (!user) {
        throw new Error("Must be authenticated to create pins");
      }

      // Create the Location protobuf message
      const location = create(LocationSchema, {
        latitude: data.location.latitude,
        longitude: data.location.longitude,
      });

      // Create the CreatePinRequest protobuf message
      const request = create(CreatePinRequestSchema, {
        content: data.content,
        location: location,
      });

      // Call the ConnectRPC service
      const response = await createPin(request);
      return response;
    },
    onMutate: async (data: CreatePinData) => {
      // Generate temporary ID for optimistic update
      const tempId = `temp-${Date.now()}`;
      const optimisticPin = {
        id: tempId,
        content: data.content,
        location: data.location,
        authorId: user!.id,
        authorUsername: user!.username,
        createdAt: new Date().toISOString(),
        commentCount: 0,
        isPending: true,
      };

      // Add optimistic pin
      addOptimisticPin(optimisticPin);
      
      return { tempId, optimisticPin };
    },
    onSuccess: (response, variables, context) => {
      if (response.pin && context) {
        // Remove optimistic pin and add real pin
        removeOptimisticPin(context.tempId);
        
        // Convert protobuf Pin to our frontend Pin type
        const pin = {
          id: response.pin.id,
          content: response.pin.content,
          location: {
            latitude: response.pin.location?.latitude || 0,
            longitude: response.pin.location?.longitude || 0,
          },
          authorId: response.pin.userId,
          authorUsername: response.pin.author?.username || user?.username || "",
          createdAt: new Date(Number(response.pin.createdAt) * 1000).toISOString(),
          commentCount: response.pin.commentCount,
          isPending: false,
        };
        
        addPin(pin);
      }
    },
    onError: (error, variables, context) => {
      // Remove optimistic pin on error
      if (context) {
        removeOptimisticPin(context.tempId);
      }
    },
  });

  return {
    createPin: mutation.mutate,
    createPinAsync: mutation.mutateAsync,
    isLoading: mutation.isPending,
    error: mutation.error,
    reset: mutation.reset,
  };
}