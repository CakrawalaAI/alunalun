import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAuthStore } from "@/features/auth/store/authStore";
import { getPin, addComment } from "@/common/services/connectrpc";
import { create } from "@bufbuild/protobuf";
import { 
  GetPinRequestSchema,
  AddCommentRequestSchema 
} from "@/common/services/connectrpc/v1/service/pin_service_pb";
import type { Comment } from "../components/comment-list";

export function useComments(pinId?: string) {
  const queryClient = useQueryClient();
  const { user } = useAuthStore();

  // Query to get pin with comments
  const { data: pinData, isLoading } = useQuery({
    queryKey: ["pin", pinId],
    queryFn: async () => {
      if (!pinId) return null;

      const request = create(GetPinRequestSchema, {
        pinId: pinId,
      });

      const response = await getPin(request);
      return response;
    },
    enabled: !!pinId,
    staleTime: 1000 * 60 * 2, // 2 minutes
  });

  // Convert protobuf comments to frontend Comment type
  const comments: Comment[] = (pinData?.comments || []).map((comment) => ({
    id: comment.id,
    content: comment.content,
    authorUsername: comment.author?.username || "",
    createdAt: new Date(Number(comment.createdAt) * 1000).toISOString(),
  }));

  // Mutation to add a comment
  const addCommentMutation = useMutation({
    mutationFn: async (content: string) => {
      if (!user || !pinId) {
        throw new Error("Must be authenticated and have a pin selected");
      }

      const request = create(AddCommentRequestSchema, {
        pinId: pinId,
        content: content,
      });

      const response = await addComment(request);
      return response;
    },
    onSuccess: () => {
      // Invalidate and refetch the pin data to get updated comments
      queryClient.invalidateQueries({ queryKey: ["pin", pinId] });
    },
  });

  const addCommentFn = async (content: string) => {
    await addCommentMutation.mutateAsync(content);
  };

  return {
    comments,
    isLoadingComments: isLoading,
    addComment: addCommentFn,
    isAddingComment: addCommentMutation.isPending,
    addCommentError: addCommentMutation.error,
  };
}