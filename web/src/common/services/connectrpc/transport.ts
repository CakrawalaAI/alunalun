import { createConnectTransport } from "@connectrpc/connect-web";
import { useAuthStore } from "@/features/auth/store/authStore";

// Create the ConnectRPC transport
export const transport = createConnectTransport({
  baseUrl: import.meta.env.VITE_API_URL || "http://localhost:8080",
  interceptors: [
    (next) => async (req) => {
      // Add authentication header if we have a token
      const { token } = useAuthStore.getState();
      if (token) {
        req.header.set("Authorization", `Bearer ${token}`);
      }
      
      return await next(req);
    },
  ],
});