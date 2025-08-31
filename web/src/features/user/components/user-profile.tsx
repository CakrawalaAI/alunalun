import { useState } from "react";
import { useAuth } from "@/features/auth/hooks";
import { useUpdateProfile } from "@/features/user/hooks";
import { UserAvatar } from "./user-avatar";

export const UserProfile = () => {
  const { user, username, isAuthenticated, isAnonymous, signOut } = useAuth();
  const updateProfile = useUpdateProfile();

  const [isEditing, setIsEditing] = useState(false);
  const [displayName, setDisplayName] = useState(user?.displayName || "");
  const [avatarUrl, setAvatarUrl] = useState(user?.avatarUrl || "");

  const handleSave = async () => {
    if (!isAuthenticated()) {
      return;
    }

    try {
      await updateProfile.mutateAsync({
        displayName: displayName || undefined,
        avatarUrl: avatarUrl || undefined,
      });
      setIsEditing(false);
    } catch (error) {
      console.error("Failed to update profile:", error);
    }
  };

  const handleCancel = () => {
    setDisplayName(user?.displayName || "");
    setAvatarUrl(user?.avatarUrl || "");
    setIsEditing(false);
  };

  return (
    <div className="rounded-lg bg-white p-6 shadow">
      <div className="flex items-start gap-4">
        <UserAvatar size="lg" />

        <div className="flex-1">
          {isEditing ? (
            <div className="space-y-4">
              <div>
                <label className="mb-1 block font-medium text-gray-700 text-sm">
                  Display Name
                </label>
                <input
                  type="text"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  placeholder="Enter display name"
                  className="w-full rounded-md border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="mb-1 block font-medium text-gray-700 text-sm">
                  Avatar URL
                </label>
                <input
                  type="url"
                  value={avatarUrl}
                  onChange={(e) => setAvatarUrl(e.target.value)}
                  placeholder="Enter avatar URL"
                  className="w-full rounded-md border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div className="flex gap-2">
                <button
                  onClick={handleSave}
                  disabled={updateProfile.isPending}
                  className="rounded-md bg-blue-500 px-4 py-2 text-white hover:bg-blue-600 disabled:opacity-50"
                >
                  {updateProfile.isPending ? "Saving..." : "Save"}
                </button>
                <button
                  onClick={handleCancel}
                  disabled={updateProfile.isPending}
                  className="rounded-md bg-gray-200 px-4 py-2 text-gray-700 hover:bg-gray-300"
                >
                  Cancel
                </button>
              </div>
            </div>
          ) : (
            <>
              <div className="mb-2">
                <h2 className="font-semibold text-xl">
                  {user?.displayName || username || "Anonymous User"}
                </h2>
                <p className="text-gray-500">@{username || "anonymous"}</p>
              </div>

              {isAuthenticated() && (
                <p className="mb-4 text-gray-600 text-sm">{user?.email}</p>
              )}

              {isAnonymous() && (
                <div className="mb-4 rounded-md border border-yellow-200 bg-yellow-50 p-3">
                  <p className="text-sm text-yellow-800">
                    You're using a temporary account. Sign in with Google to
                    save your content permanently.
                  </p>
                </div>
              )}

              <div className="flex gap-2">
                {isAuthenticated() && (
                  <button
                    onClick={() => setIsEditing(true)}
                    className="rounded-md bg-blue-500 px-4 py-2 text-white hover:bg-blue-600"
                  >
                    Edit Profile
                  </button>
                )}

                <button
                  onClick={signOut}
                  className="rounded-md bg-red-500 px-4 py-2 text-white hover:bg-red-600"
                >
                  Sign Out
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};
