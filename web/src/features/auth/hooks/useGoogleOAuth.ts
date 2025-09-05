import { useMutation } from "@tanstack/react-query";
import { useAuthStore } from "../store/authStore";
import { authenticate } from "@/common/services/connectrpc";
import { create } from "@bufbuild/protobuf";
import { AuthenticateRequestSchema } from "@/common/services/connectrpc/v1/service/auth_pb";

export function useGoogleOAuth() {
  const { setAuth } = useAuthStore();

  const authenticateMutation = useMutation({
    mutationFn: async (googleIdToken: string) => {
      const request = create(AuthenticateRequestSchema, {
        provider: "google",
        credential: googleIdToken,
        // sessionId can be added later for session migration
      });

      const response = await authenticate(request);
      return response;
    },
    onSuccess: (response) => {
      if (response.user) {
        // Convert protobuf User to frontend User type
        const user = {
          id: response.user.id,
          username: response.user.username,
          email: response.user.email || undefined,
          picture: response.user.metadata?.["profile_picture_url"] || undefined,
          createdAt: new Date(Number(response.user.createdAt) * 1000).toISOString(),
        };
        
        setAuth(user, response.token);
      }
    },
    onError: (error) => {
      console.error("Google authentication failed:", error);
      throw error;
    },
  });

  const signInWithGoogle = async () => {
    // TODO: Integrate with Google Identity Services (GIS) or Google OAuth
    // For now, this is a placeholder that expects the Google ID token
    // Real implementation would use @google-cloud/identity or google-auth-library
    
    // This is still a development placeholder until we integrate actual Google OAuth
    throw new Error("Google OAuth integration not yet implemented. Please add Google Identity Services integration.");
    
    // Real implementation would look like:
    // 1. Load Google Identity Services
    // 2. Initialize Google OAuth client
    // 3. Get ID token from Google
    // 4. Call our authenticate endpoint with the ID token
    // const idToken = await getGoogleIdToken();
    // await authenticateMutation.mutateAsync(idToken);
  };

  return {
    signInWithGoogle,
    isLoading: authenticateMutation.isPending,
    error: authenticateMutation.error,
  };
}