import { useAuthStore } from "@/features/auth/store/authStore";

/**
 * Username Display - shows current user's username in top-left corner
 */
export function UsernameDisplay() {
  const { user } = useAuthStore();

  if (!user?.username) {
    return null;
  }

  return (
    <div 
      className="pointer-events-auto fixed top-4 left-4 text-gray-500 text-xs"
      style={{ zIndex: 100 }} // Higher than tap zones
    >
      {user.username}
    </div>
  );
}