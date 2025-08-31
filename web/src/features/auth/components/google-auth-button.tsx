import { useCallback, useEffect } from "react";
import { useAuthenticate } from "@/features/auth/hooks";

interface GoogleAuthButtonProps {
  onSuccess?: () => void;
  onError?: (error: Error) => void;
}

declare global {
  interface Window {
    google?: {
      accounts: {
        id: {
          initialize: (config: any) => void;
          renderButton: (element: HTMLElement, config: any) => void;
          prompt: () => void;
        };
      };
    };
  }
}

export const GoogleAuthButton = ({ onSuccess, onError }: GoogleAuthButtonProps) => {
  const authenticate = useAuthenticate();
  
  const handleCredentialResponse = useCallback(async (response: any) => {
    try {
      // The response.credential is the Google ID token
      await authenticate.mutateAsync({
        provider: 'google',
        credential: response.credential,
      });
      
      onSuccess?.();
    } catch (error) {
      console.error('Google authentication failed:', error);
      onError?.(error as Error);
    }
  }, [authenticate, onSuccess, onError]);
  
  useEffect(() => {
    // Load Google Sign-In script
    const script = document.createElement('script');
    script.src = 'https://accounts.google.com/gsi/client';
    script.async = true;
    script.defer = true;
    document.body.appendChild(script);
    
    script.onload = () => {
      // Initialize Google Sign-In
      if (window.google) {
        window.google.accounts.id.initialize({
          client_id: import.meta.env.VITE_GOOGLE_CLIENT_ID,
          callback: handleCredentialResponse,
          auto_select: false,
          cancel_on_tap_outside: true,
        });
        
        // Render the button
        const buttonElement = document.getElementById('google-signin-button');
        if (buttonElement) {
          window.google.accounts.id.renderButton(buttonElement, {
            type: 'standard',
            shape: 'rectangular',
            theme: 'outline',
            text: 'signin_with',
            size: 'large',
            width: '100%',
          });
        }
      }
    };
    
    return () => {
      // Cleanup
      document.body.removeChild(script);
    };
  }, [handleCredentialResponse]);
  
  return (
    <div id="google-signin-button" className="w-full" />
  );
};