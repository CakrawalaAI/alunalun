import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/common/components/ui/dialog";
import { GoogleSignInButton } from "./google-sign-in-button";
import { UsernamePickerForm } from "./username-picker-form";
import { TurnstileWidget, useTurnstile } from "./turnstile-widget";
import { useAuthModal } from "../hooks/useAuthModal";
import { useAuthStore } from "../store/authStore";

export function AuthModal() {
  const { isOpen, closeModal } = useAuthModal();
  const { user } = useAuthStore();
  const [showUsernamePicker, setShowUsernamePicker] = useState(false);
  const turnstile = useTurnstile();

  // Show username picker if user exists but has no username
  const needsUsername = user && !user.username;

  // Get Turnstile site key from environment
  const turnstileSiteKey = import.meta.env.VITE_TURNSTILE_SITE_KEY;

  return (
    <Dialog open={isOpen} onOpenChange={closeModal}>
      <DialogContent className="sm:max-w-md">
        {needsUsername || showUsernamePicker ? (
          <>
            <DialogHeader>
              <DialogTitle>Choose your username</DialogTitle>
            </DialogHeader>
            <UsernamePickerForm
              onComplete={() => {
                setShowUsernamePicker(false);
                closeModal();
              }}
              onCancel={() => {
                setShowUsernamePicker(false);
                if (!needsUsername) closeModal();
              }}
            />
          </>
        ) : (
          <>
            <DialogHeader>
              <DialogTitle>Continue with Google</DialogTitle>
            </DialogHeader>
            <div className="flex flex-col gap-4 py-4">
              <GoogleSignInButton
                turnstileToken={turnstile.token}
                turnstileVerified={turnstile.isVerified}
                onSuccess={() => {
                  // Check if username is needed after auth
                  if (needsUsername) {
                    setShowUsernamePicker(true);
                  } else {
                    closeModal();
                  }
                }}
              />
              
              {turnstileSiteKey && (
                <div className="mt-4">
                  <TurnstileWidget
                    siteKey={turnstileSiteKey}
                    onSuccess={turnstile.handleSuccess}
                    onError={turnstile.handleError}
                    onExpired={turnstile.handleExpired}
                    theme="auto"
                    size="normal"
                    className="flex justify-center"
                  />
                  {turnstile.error && (
                    <p className="text-sm text-red-600 mt-2 text-center">
                      {turnstile.error}
                    </p>
                  )}
                </div>
              )}

              {!turnstileSiteKey && (
                <div className="text-sm text-gray-500 text-center">
                  Security verification disabled in development
                </div>
              )}
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}