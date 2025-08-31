import { useState, useEffect } from "react";
import { useCheckUsername, useInitAnonymous } from "@/features/auth/hooks";

interface UsernameModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export const UsernameModal = ({ isOpen, onClose, onSuccess }: UsernameModalProps) => {
  const [username, setUsername] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isChecking, setIsChecking] = useState(false);
  
  const checkUsername = useCheckUsername();
  const initAnonymous = useInitAnonymous();
  
  // Reset state when modal opens/closes
  useEffect(() => {
    if (!isOpen) {
      setUsername("");
      setError(null);
      setIsChecking(false);
    }
  }, [isOpen]);
  
  // Validate username format
  const validateUsername = (value: string): boolean => {
    const pattern = /^[a-zA-Z0-9_]{3,100}$/;
    return pattern.test(value);
  };
  
  const handleSubmit = async (e?: React.FormEvent) => {
    e?.preventDefault();
    
    // Validate format first
    if (!validateUsername(username)) {
      setError("Username must be 3-100 characters, letters, numbers, and underscores only");
      return;
    }
    
    setIsChecking(true);
    setError(null);
    
    try {
      // Check if username is available
      const checkResult = await checkUsername.mutateAsync(username);
      
      if (!checkResult.available) {
        setError(checkResult.message || "Username is already taken");
        setIsChecking(false);
        return;
      }
      
      // Create anonymous session
      await initAnonymous.mutateAsync(username);
      
      // Success!
      onSuccess();
      onClose();
    } catch (err) {
      setError("Failed to create session. Please try again.");
      console.error('Username modal error:', err);
    } finally {
      setIsChecking(false);
    }
  };
  
  if (!isOpen) return null;
  
  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
        <div className="mb-4">
          <h2 className="text-2xl font-bold mb-2">Choose your username</h2>
          <p className="text-gray-600">
            Start posting immediately with just a username. 
            You can verify your account later to keep your content.
          </p>
        </div>
        
        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <input
              type="text"
              value={username}
              onChange={(e) => {
                setUsername(e.target.value);
                setError(null);
              }}
              placeholder="Enter username"
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              autoFocus
              disabled={isChecking}
              maxLength={100}
            />
            {error && (
              <p className="mt-2 text-sm text-red-600">{error}</p>
            )}
            <p className="mt-2 text-xs text-gray-500">
              3-100 characters, letters, numbers, and underscores only
            </p>
          </div>
          
          <div className="flex gap-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
              disabled={isChecking}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="flex-1 px-4 py-2 text-white bg-blue-500 rounded-md hover:bg-blue-600 transition-colors disabled:opacity-50"
              disabled={!username || isChecking}
            >
              {isChecking ? "Checking..." : "Continue without logging in"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};