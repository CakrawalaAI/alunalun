interface JWTPayload {
  type: "anonymous" | "authenticated";
  session_id?: string;
  sub?: string;
  username?: string;
  iat: number;
  exp?: number;
  refresh_until?: number;
}

/**
 * Decode JWT token without verification (client-side only)
 */
export function decodeToken(token: string): JWTPayload | null {
  try {
    const parts = token.split(".");
    if (parts.length !== 3) {
      return null;
    }

    const payload = parts[1];
    const decoded = atob(payload.replace(/-/g, "+").replace(/_/g, "/"));
    return JSON.parse(decoded);
  } catch (error) {
    console.error("Failed to decode token:", error);
    return null;
  }
}

/**
 * Check if token is expiring soon (within 5 minutes)
 */
export function isTokenExpiringSoon(token: string): boolean {
  const payload = decodeToken(token);
  if (!(payload && payload.exp)) {
    // Anonymous tokens don't expire
    return false;
  }

  const expiryTime = payload.exp * 1000; // Convert to milliseconds
  const now = Date.now();
  const fiveMinutes = 5 * 60 * 1000;

  return expiryTime - now < fiveMinutes;
}

/**
 * Check if token is expired
 */
export function isTokenExpired(token: string): boolean {
  const payload = decodeToken(token);
  if (!(payload && payload.exp)) {
    // Anonymous tokens don't expire
    return false;
  }

  const expiryTime = payload.exp * 1000; // Convert to milliseconds
  const now = Date.now();

  return now > expiryTime;
}

/**
 * Check if token can be refreshed
 */
export function canRefreshToken(token: string): boolean {
  const payload = decodeToken(token);
  if (!payload || payload.type !== "authenticated") {
    return false;
  }

  if (!payload.refresh_until) {
    return false;
  }

  const refreshUntil = payload.refresh_until * 1000; // Convert to milliseconds
  const now = Date.now();

  return now < refreshUntil;
}

/**
 * Get token type
 */
export function getTokenType(
  token: string,
): "anonymous" | "authenticated" | null {
  const payload = decodeToken(token);
  return payload?.type || null;
}
