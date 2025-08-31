import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { AuthService } from "@/common/services/connectrpc/v1/service/auth_connect";
import { UserService } from "@/common/services/connectrpc/v1/service/user_connect";
import { useAuthStore } from "@/features/auth/store/auth-store";
import { isTokenExpiringSoon, canRefreshToken } from "@/features/auth/lib/token-utils";

// Create transport with auth interceptor
const createAuthTransport = () => {
  return createConnectTransport({
    baseUrl: import.meta.env.VITE_API_URL || "http://localhost:8080",
    interceptors: [
      (next) => async (req) => {
        // Add auth headers from store
        const headers = useAuthStore.getState().getAuthHeaders();
        Object.entries(headers).forEach(([key, value]) => {
          req.header.set(key, value);
        });
        
        // Auto-refresh logic for authenticated tokens
        const authType = useAuthStore.getState().type;
        const token = useAuthStore.getState().token;
        
        if (authType === 'authenticated' && token) {
          // Check if token is expiring soon and can be refreshed
          if (isTokenExpiringSoon(token) && canRefreshToken(token)) {
            try {
              // Refresh the token before making the request
              const response = await authClient.refreshToken({
                expiredToken: token,
              });
              
              // Update the store with new token
              const user = useAuthStore.getState().user;
              if (user) {
                useAuthStore.getState().setAuthenticatedAuth(response.token, user);
                // Update the request header with new token
                req.header.set('Authorization', `Bearer ${response.token}`);
              }
            } catch (error) {
              console.error('Failed to refresh token:', error);
              // Token refresh failed, let the request continue with expired token
              // The server will return 401 and we'll handle it in the UI
            }
          }
        }
        
        return next(req);
      },
    ],
  });
};

// Create singleton transport instance
let transport: ReturnType<typeof createAuthTransport> | null = null;

const getTransport = () => {
  if (!transport) {
    transport = createAuthTransport();
  }
  return transport;
};

// Export client instances
export const authClient = createPromiseClient(AuthService, getTransport());
export const userClient = createPromiseClient(UserService, getTransport());

// Helper to recreate clients if needed (e.g., after environment change)
export const reinitializeClients = () => {
  transport = null;
  const newTransport = getTransport();
  authClient = createPromiseClient(AuthService, newTransport);
  userClient = createPromiseClient(UserService, newTransport);
};