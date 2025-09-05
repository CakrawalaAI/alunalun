import { LogIn } from "lucide-react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/common/components/ui/popover";
import { cn } from "@/common/lib/utils";
import { useAuthModal } from "@/features/auth/hooks/useAuthModal";
import { useAuthStore } from "@/features/auth/store/authStore";

export function AccountButton() {
  const { user, isAuthenticated } = useAuthStore();
  const { openAuthModal } = useAuthModal();
  const { logout } = useAuthStore();

  if (!isAuthenticated) {
    return (
      <button
        type="button"
        onClick={openAuthModal}
        className={cn(
          "h-10 w-10 rounded-full bg-white/90 backdrop-blur-md",
          "border border-gray-200/60 shadow-lg",
          "flex items-center justify-center",
          "transition-colors hover:bg-gray-100/50",
          "pointer-events-auto",
        )}
        aria-label="Sign in"
      >
        <LogIn className="h-5 w-5 text-gray-700" />
      </button>
    );
  }

  // User initial button with dropdown
  const initial = user?.username?.[0]?.toUpperCase() || "U";

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          type="button"
          className={cn(
            "h-10 w-10 rounded-full bg-white/90 backdrop-blur-md",
            "border border-gray-200/60 shadow-lg",
            "flex items-center justify-center",
            "transition-colors hover:bg-gray-100/50",
            "font-semibold text-gray-700",
            "pointer-events-auto",
          )}
          aria-label="Account menu"
        >
          {initial}
        </button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-48">
        <div className="flex flex-col gap-1">
          <div className="px-2 py-1.5 font-medium text-sm">
            @{user?.username}
          </div>
          <div className="h-px bg-gray-200" />
          <button
            type="button"
            onClick={() => logout()}
            className="rounded-md px-2 py-1.5 text-left text-sm transition-colors hover:bg-gray-100"
          >
            Logout
          </button>
        </div>
      </PopoverContent>
    </Popover>
  );
}