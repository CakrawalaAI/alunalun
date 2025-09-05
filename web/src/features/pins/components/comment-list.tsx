import { formatDistanceToNow } from "@/common/utils/timestamp";

export interface Comment {
  id: string;
  content: string;
  authorUsername: string;
  createdAt: string;
}

interface CommentListProps {
  comments: Comment[];
}

export function CommentList({ comments }: CommentListProps) {
  if (comments.length === 0) {
    return (
      <div className="text-sm text-gray-500">
        No comments yet. Be the first to comment!
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {comments.map((comment) => (
        <div key={comment.id} className="text-sm">
          <div className="flex items-baseline gap-2 mb-1">
            <span className="font-medium">@{comment.authorUsername}</span>
            <span className="text-xs text-gray-500">
              {formatDistanceToNow(comment.createdAt)}
            </span>
          </div>
          <div className="text-gray-700">
            {comment.content}
          </div>
        </div>
      ))}
    </div>
  );
}