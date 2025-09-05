import { useEffect, useRef, useState } from "react";

export interface TurnstileWidgetProps {
  siteKey: string;
  onSuccess: (token: string) => void;
  onError?: (error: string) => void;
  onExpired?: () => void;
  theme?: "light" | "dark" | "auto";
  size?: "normal" | "compact";
  className?: string;
}

// Extend Window interface to include turnstile
declare global {
  interface Window {
    turnstile?: {
      render: (element: HTMLElement, options: TurnstileOptions) => string;
      remove: (widgetId: string) => void;
      reset: (widgetId: string) => void;
      getResponse: (widgetId: string) => string;
    };
    onloadTurnstileCallback?: () => void;
  }
}

interface TurnstileOptions {
  sitekey: string;
  callback: (token: string) => void;
  "error-callback"?: (error: string) => void;
  "expired-callback"?: () => void;
  theme?: "light" | "dark" | "auto";
  size?: "normal" | "compact";
}

export function TurnstileWidget({
  siteKey,
  onSuccess,
  onError,
  onExpired,
  theme = "auto",
  size = "normal",
  className,
}: TurnstileWidgetProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const widgetIdRef = useRef<string | null>(null);
  const [isLoaded, setIsLoaded] = useState(false);
  const [isRendered, setIsRendered] = useState(false);

  // Load the Turnstile script
  useEffect(() => {
    if (typeof window === "undefined") return;

    // Check if script is already loaded
    if (window.turnstile) {
      setIsLoaded(true);
      return;
    }

    // Create script element
    const script = document.createElement("script");
    script.src = "https://challenges.cloudflare.com/turnstile/v0/api.js";
    script.async = true;
    script.defer = true;

    // Set up callback for when script loads
    window.onloadTurnstileCallback = () => {
      setIsLoaded(true);
    };

    script.onload = () => {
      if (window.turnstile) {
        setIsLoaded(true);
      }
    };

    document.head.appendChild(script);

    return () => {
      // Cleanup
      if (script.parentNode) {
        script.parentNode.removeChild(script);
      }
      delete window.onloadTurnstileCallback;
    };
  }, []);

  // Render the widget when script is loaded
  useEffect(() => {
    if (!isLoaded || !containerRef.current || isRendered || !window.turnstile) {
      return;
    }

    try {
      const widgetId = window.turnstile.render(containerRef.current, {
        sitekey: siteKey,
        callback: (token: string) => {
          onSuccess(token);
        },
        "error-callback": (error: string) => {
          onError?.(error);
        },
        "expired-callback": () => {
          onExpired?.();
        },
        theme,
        size,
      });

      widgetIdRef.current = widgetId;
      setIsRendered(true);
    } catch (error) {
      console.error("Failed to render Turnstile widget:", error);
      onError?.("Failed to load security verification");
    }
  }, [isLoaded, siteKey, onSuccess, onError, onExpired, theme, size, isRendered]);

  // Cleanup widget on unmount
  useEffect(() => {
    return () => {
      if (widgetIdRef.current && window.turnstile) {
        try {
          window.turnstile.remove(widgetIdRef.current);
        } catch (error) {
          console.error("Failed to cleanup Turnstile widget:", error);
        }
      }
    };
  }, []);

  // Reset widget function
  const reset = () => {
    if (widgetIdRef.current && window.turnstile) {
      try {
        window.turnstile.reset(widgetIdRef.current);
      } catch (error) {
        console.error("Failed to reset Turnstile widget:", error);
      }
    }
  };

  // Get response function  
  const getResponse = () => {
    if (widgetIdRef.current && window.turnstile) {
      try {
        return window.turnstile.getResponse(widgetIdRef.current);
      } catch (error) {
        console.error("Failed to get Turnstile response:", error);
        return null;
      }
    }
    return null;
  };

  // Expose reset and getResponse methods via ref
  useEffect(() => {
    if (containerRef.current) {
      (containerRef.current as any).reset = reset;
      (containerRef.current as any).getResponse = getResponse;
    }
  }, []);

  if (!isLoaded) {
    return (
      <div className={`flex items-center justify-center p-4 ${className || ""}`}>
        <div className="text-sm text-gray-500">Loading security verification...</div>
      </div>
    );
  }

  return (
    <div className={className}>
      <div ref={containerRef} />
    </div>
  );
}

// Hook for using Turnstile widget
export function useTurnstile() {
  const [token, setToken] = useState<string | null>(null);
  const [isVerified, setIsVerified] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSuccess = (token: string) => {
    setToken(token);
    setIsVerified(true);
    setError(null);
  };

  const handleError = (error: string) => {
    setError(error);
    setIsVerified(false);
    setToken(null);
  };

  const handleExpired = () => {
    setToken(null);
    setIsVerified(false);
    setError("Verification expired. Please try again.");
  };

  const reset = () => {
    setToken(null);
    setIsVerified(false);
    setError(null);
  };

  return {
    token,
    isVerified,
    error,
    handleSuccess,
    handleError,
    handleExpired,
    reset,
  };
}