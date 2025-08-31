export interface AuthState {
  type: 'none' | 'anonymous' | 'authenticated';
  token: string | null;
  sessionId: string | null;
  username: string | null;
  user: User | null;
}

export interface User {
  id: string;
  email: string;
  username: string;
  displayName?: string;
  avatarUrl?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CheckUsernameResponse {
  available: boolean;
  message?: string;
}

export interface InitAnonymousResponse {
  token: string;
  sessionId: string;
  username: string;
}

export interface AuthenticateResponse {
  token: string;
  user: User;
  sessionMigrated: boolean;
}

export interface RefreshTokenResponse {
  token: string;
}