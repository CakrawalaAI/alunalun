import { useAuth } from "@/features/auth/hooks";

interface UserAvatarProps {
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export const UserAvatar = ({ size = 'md', className = '' }: UserAvatarProps) => {
  const { user, username, isAuthenticated } = useAuth();
  
  const sizeClasses = {
    sm: 'w-8 h-8 text-sm',
    md: 'w-10 h-10 text-base',
    lg: 'w-16 h-16 text-xl',
  };
  
  const avatarUrl = user?.avatarUrl;
  const displayName = user?.displayName || username || 'A';
  const initial = displayName.charAt(0).toUpperCase();
  
  if (avatarUrl) {
    return (
      <img
        src={avatarUrl}
        alt={displayName}
        className={`rounded-full object-cover ${sizeClasses[size]} ${className}`}
      />
    );
  }
  
  // Generate a color based on the username for consistency
  const colors = [
    'bg-red-500',
    'bg-orange-500',
    'bg-yellow-500',
    'bg-green-500',
    'bg-teal-500',
    'bg-blue-500',
    'bg-indigo-500',
    'bg-purple-500',
    'bg-pink-500',
  ];
  
  const colorIndex = (username || 'anonymous').charCodeAt(0) % colors.length;
  const bgColor = isAuthenticated() ? colors[colorIndex] : 'bg-gray-400';
  
  return (
    <div
      className={`rounded-full flex items-center justify-center text-white font-semibold ${bgColor} ${sizeClasses[size]} ${className}`}
    >
      {initial}
    </div>
  );
};