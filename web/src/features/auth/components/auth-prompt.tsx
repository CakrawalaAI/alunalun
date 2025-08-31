import { useState } from "react";
import { useAuth } from "@/features/auth/hooks";
import { GoogleAuthButton } from "./google-auth-button";
import { UsernameModal } from "./username-modal";

interface AuthPromptProps {
  message?: string;
  onSuccess?: () => void;
  onCancel?: () => void;
}

export const AuthPrompt = ({
  message = "Sign in to continue",
  onSuccess,
  onCancel,
}: AuthPromptProps) => {
  const [showUsernameModal, setShowUsernameModal] = useState(false);
  const { hasAuth } = useAuth();

  // If user already has auth, don't show the prompt
  if (hasAuth()) {
    onSuccess?.();
    return null;
  }

  return (
    <>
      <div className="w-full max-w-md rounded-lg bg-white p-6">
        <h2 className="mb-4 font-bold text-xl">{message}</h2>

        <div className="space-y-3">
          {/* Google Sign In */}
          <GoogleAuthButton
            onSuccess={() => {
              onSuccess?.();
            }}
          />

          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-gray-300 border-t" />
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="bg-white px-2 text-gray-500">or</span>
            </div>
          </div>

          {/* Continue without account */}
          <button
            onClick={() => setShowUsernameModal(true)}
            className="w-full rounded-md bg-gray-100 px-4 py-2 text-gray-700 transition-colors hover:bg-gray-200"
          >
            Continue without an account
          </button>

          {onCancel && (
            <button
              onClick={onCancel}
              className="w-full px-4 py-2 text-gray-500 transition-colors hover:text-gray-700"
            >
              Cancel
            </button>
          )}
        </div>
      </div>

      <UsernameModal
        isOpen={showUsernameModal}
        onClose={() => setShowUsernameModal(false)}
        onSuccess={() => {
          setShowUsernameModal(false);
          onSuccess?.();
        }}
      />
    </>
  );
};
