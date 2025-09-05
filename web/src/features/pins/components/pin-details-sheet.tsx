import {
  Sheet,
  SheetContent,
  SheetHeader,
} from "@/common/components/ui/sheet";
import { Input } from "@/common/components/ui/input";
import { Button } from "@/common/components/ui/button";
import { MessageCircle } from "lucide-react";
import { usePinsStore } from "../store/pinsStore";
import { useComments } from "../hooks/useComments";
import { CommentList } from "./comment-list";
import { useState } from "react";
import { useAuthStore } from "@/features/auth/store/authStore";
import { formatDistanceToNow } from "@/common/utils/timestamp";

export function PinDetailsSheet() {
  const { selectedPin, selectPin } = usePinsStore();
  const { comments, addComment, isLoadingComments } = useComments(selectedPin?.id);
  const { isAuthenticated } = useAuthStore();
  const [newComment, setNewComment] = useState("");
  const [isAddingComment, setIsAddingComment] = useState(false);

  const handleAddComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newComment.trim() || !selectedPin) return;

    setIsAddingComment(true);
    try {
      await addComment(newComment.trim());
      setNewComment("");
    } catch (error) {
      console.error("Failed to add comment:", error);
    } finally {
      setIsAddingComment(false);
    }
  };

  if (!selectedPin) return null;

  return (
    <Sheet open={!!selectedPin} onOpenChange={(open) => !open && selectPin(null)}>
      <SheetContent 
        side="bottom"
        className="sm:max-w-md mx-auto rounded-t-xl h-[70vh] flex flex-col"
      >
        <SheetHeader>
          <div className="flex items-center justify-between">
            <div className="text-sm text-gray-600">
              @{selectedPin.authorUsername} · {formatDistanceToNow(selectedPin.createdAt)}
            </div>
          </div>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto py-4">
          <div className="space-y-4">
            {/* Pin content */}
            <div className="text-lg">
              {selectedPin.content}
            </div>

            {/* Comments section */}
            <div className="border-t pt-4">
              <div className="flex items-center gap-2 mb-4">
                <MessageCircle className="w-4 h-4 text-gray-500" />
                <span className="text-sm text-gray-600">
                  {comments.length} comment{comments.length !== 1 ? "s" : ""}
                </span>
              </div>

              {isLoadingComments ? (
                <div className="text-sm text-gray-500">Loading comments...</div>
              ) : (
                <CommentList comments={comments} />
              )}
            </div>
          </div>
        </div>

        {/* Add comment form */}
        {isAuthenticated && (
          <form onSubmit={handleAddComment} className="border-t pt-4 mt-auto">
            <div className="flex gap-2">
              <Input
                value={newComment}
                onChange={(e) => setNewComment(e.target.value)}
                placeholder="Add a comment..."
                disabled={isAddingComment}
                className="flex-1"
              />
              <Button
                type="submit"
                size="sm"
                disabled={!newComment.trim() || isAddingComment}
              >
                Post
              </Button>
            </div>
          </form>
        )}
      </SheetContent>
    </Sheet>
  );
}