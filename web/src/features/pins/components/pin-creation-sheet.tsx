import { useState } from "react";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from "@/common/components/ui/sheet";
import { Textarea } from "@/common/components/ui/textarea";
import { Button } from "@/common/components/ui/button";
import { usePinsStore } from "../store/pinsStore";
import { useCreatePin } from "../hooks/useCreatePin";
import { cn } from "@/common/lib/utils";

export function PinCreationSheet() {
  const { isCreating, creatingLocation, setCreating } = usePinsStore();
  const { createPin, isLoading } = useCreatePin();
  const [content, setContent] = useState("");
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!content.trim() || content.length < 10) {
      setError("Content must be at least 10 characters");
      return;
    }

    if (!creatingLocation) {
      setError("Location not set");
      return;
    }

    try {
      await createPin({
        content: content.trim(),
        location: {
          latitude: creatingLocation.lat,
          longitude: creatingLocation.lng,
        },
      });
      
      // Reset form and close
      setContent("");
      setError(null);
      setCreating(false);
    } catch (error) {
      setError("Failed to create pin. Please try again.");
    }
  };

  const handleCancel = () => {
    setContent("");
    setError(null);
    setCreating(false);
  };

  return (
    <Sheet open={isCreating} onOpenChange={(open) => !open && handleCancel()}>
      <SheetContent 
        side="bottom" 
        className="sm:max-w-md mx-auto rounded-t-xl"
      >
        <SheetHeader>
          <SheetTitle>What happened here?</SheetTitle>
        </SheetHeader>
        
        <form onSubmit={handleSubmit} className="mt-4 space-y-4">
          <div>
            <Textarea
              value={content}
              onChange={(e) => {
                setContent(e.target.value);
                setError(null);
              }}
              placeholder="Share what makes this place special..."
              className="min-h-[100px] resize-none"
              maxLength={280}
              required
            />
            <div className="mt-1 flex justify-between text-xs text-gray-500">
              <span>{error && <span className="text-red-500">{error}</span>}</span>
              <span className={cn(
                content.length > 260 && "text-orange-500",
                content.length === 280 && "text-red-500"
              )}>
                {content.length}/280
              </span>
            </div>
          </div>

          <div className="flex gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={handleCancel}
              className="flex-1"
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              className="flex-1"
              disabled={isLoading || content.length < 10}
            >
              {isLoading ? "Creating..." : "Drop Pin →"}
            </Button>
          </div>
        </form>
      </SheetContent>
    </Sheet>
  );
}