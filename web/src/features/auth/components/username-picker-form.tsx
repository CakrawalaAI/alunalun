import { useState, useEffect } from "react";
import { Input } from "@/common/components/ui/input";
import { Button } from "@/common/components/ui/button";
import { useUsernameValidation } from "../hooks/useUsernameValidation";
import { useAuthStore } from "../store/authStore";
import { cn } from "@/common/lib/utils";

interface UsernamePickerFormProps {
  onComplete: () => void;
  onCancel: () => void;
}

export function UsernamePickerForm({ onComplete, onCancel }: UsernamePickerFormProps) {
  const [username, setUsername] = useState("");
  const { isValidating, isAvailable, validate } = useUsernameValidation();
  const { updateUser } = useAuthStore();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (username.length >= 3) {
      const timer = setTimeout(() => {
        validate(username);
      }, 300);
      return () => clearTimeout(timer);
    }
  }, [username, validate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!isAvailable || username.length < 3) {
      setError("Username must be at least 3 characters and available");
      return;
    }

    try {
      // Update username in local store (backend will be updated on next API call with auth token)
      updateUser({ username });
      onComplete();
    } catch (error) {
      setError("Failed to set username. Please try again.");
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label htmlFor="username" className="text-sm font-medium">
          Username
        </label>
        <div className="relative mt-1">
          <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">
            @
          </span>
          <Input
            id="username"
            value={username}
            onChange={(e) => {
              setUsername(e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, ""));
              setError(null);
            }}
            className="pl-8"
            placeholder="johndoe123"
            required
            minLength={3}
            maxLength={20}
          />
        </div>
        {username.length >= 3 && (
          <p className={cn(
            "text-sm mt-1",
            isValidating && "text-gray-500",
            isAvailable === true && "text-green-600",
            isAvailable === false && "text-red-600"
          )}>
            {isValidating
              ? "Checking..."
              : isAvailable
              ? "✓ Username available"
              : "✗ Username taken"}
          </p>
        )}
        {error && (
          <p className="text-sm text-red-600 mt-1">{error}</p>
        )}
      </div>

      <div className="flex gap-2">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          className="flex-1"
        >
          Cancel
        </Button>
        <Button
          type="submit"
          disabled={!isAvailable || isValidating || username.length < 3}
          className="flex-1"
        >
          Continue →
        </Button>
      </div>
    </form>
  );
}