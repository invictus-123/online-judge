import { createContext } from 'react';
import type { LoginRequest, RegisterRequest } from '../types/api';
import { UserRole } from '../types/api';

interface User {
  id: string;
  handle: string;
  email: string;
  firstName: string;
  lastName: string;
  role: UserRole;
}

interface AuthState {
  user: User | null;
  token: string | null;
  isInitializing: boolean;
  isAuthenticating: boolean;
  isAuthenticated: boolean;
}

interface AuthContextValue extends AuthState {
  login: (credentials: LoginRequest) => Promise<void>;
  register: (userData: RegisterRequest) => Promise<void>;
  logout: () => Promise<void>;
  checkAuth: () => void;
}

export const AuthContext = createContext<AuthContextValue | null>(null);

export type { User, AuthState, AuthContextValue };