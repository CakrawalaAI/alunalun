import { useState } from "react";
import { UsernameModal } from "./username-modal";
import { GoogleAuthButton } from "./google-auth-button";
import { useAuth } from "@/features/auth/hooks";

interface AuthPromptProps {
  message?: string;
  onSuccess?: () => void;
  onCancel?: () => void;
}

export const AuthPrompt = ({ 
  message = "Sign in to continue", 
  onSuccess,
  onCancel 
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
      <div className="bg-white rounded-lg p-6 max-w-md w-full">
        <h2 className="text-xl font-bold mb-4">{message}</h2>
        
        <div className="space-y-3">
          {/* Google Sign In */}
          <GoogleAuthButton 
            onSuccess={() => {
              onSuccess?.();
            }}
          />
          
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-gray-300" />
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-2 bg-white text-gray-500">or</span>
            </div>
          </div>
          
          {/* Continue without account */}
          <button
            onClick={() => setShowUsernameModal(true)}
            className="w-full px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
          >
            Continue without an account
          </button>
          
          {onCancel && (
            <button
              onClick={onCancel}
              className="w-full px-4 py-2 text-gray-500 hover:text-gray-700 transition-colors"
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