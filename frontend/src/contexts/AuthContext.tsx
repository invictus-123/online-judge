import type { ReactNode } from 'react';
import { useReducer, useEffect, useMemo, useCallback } from 'react';
import { jwtDecode } from 'jwt-decode';
import { authService, tokenStorage } from '../services';
import type { LoginRequest, RegisterRequest } from '../types/api';
import { UserRole } from '../types/api';
import type { User, AuthState, AuthContextValue } from './auth-context';
import { AuthContext } from './auth-context';
import toast from 'react-hot-toast';

type AuthAction =
  | { type: 'SET_INITIALIZING'; payload: boolean }
  | { type: 'SET_AUTHENTICATING'; payload: boolean }
  | { type: 'SET_USER'; payload: { user: User; token: string } }
  | { type: 'CLEAR_USER' }
  | { type: 'UPDATE_USER'; payload: Partial<User> };

const initialState: AuthState = {
  user: null,
  token: null,
  isInitializing: true,
  isAuthenticating: false,
  isAuthenticated: false,
};

function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case 'SET_INITIALIZING':
      return { ...state, isInitializing: action.payload };
    case 'SET_AUTHENTICATING':
      return { ...state, isAuthenticating: action.payload };
    case 'SET_USER':
      return {
        ...state,
        user: action.payload.user,
        token: action.payload.token,
        isAuthenticated: true,
        isInitializing: false,
        isAuthenticating: false,
      };
    case 'CLEAR_USER':
      return {
        ...state,
        user: null,
        token: null,
        isAuthenticated: false,
        isInitializing: false,
        isAuthenticating: false,
      };
    case 'UPDATE_USER':
      return {
        ...state,
        user: state.user ? { ...state.user, ...action.payload } : null,
      };
    default:
      return state;
  }
}

interface JwtPayload {
  sub: string; // handle
  userId: string;
  email: string;
  firstName: string;
  lastName: string;
  role: UserRole;
  iat: number;
  exp: number;
}

interface AuthProviderProps {
  children: ReactNode;
}

const isTokenValid = (token: string): boolean => {
  try {
    const decoded: JwtPayload = jwtDecode(token);
    const now = Date.now() / 1000;
    return decoded.exp > now;
  } catch {
    return false;
  }
};

const getUserFromToken = (token: string): User | null => {
  try {
    const decoded: JwtPayload = jwtDecode(token);
    return {
      id: decoded.userId,
      handle: decoded.sub,
      email: decoded.email,
      firstName: decoded.firstName,
      lastName: decoded.lastName,
      role: decoded.role,
    };
  } catch (error) {
    console.error('Failed to decode token:', error);
    return null;
  }
};

export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [state, dispatch] = useReducer(authReducer, initialState);

  const checkAuth = useCallback(() => {
    const token = tokenStorage.get();
    
    if (token && isTokenValid(token)) {
      const user = getUserFromToken(token);
      if (user) {
        dispatch({ type: 'SET_USER', payload: { user, token } });
        return;
      }
    }

    tokenStorage.remove();
    dispatch({ type: 'SET_INITIALIZING', payload: false });
  }, []);

  const login = useCallback(async (credentials: LoginRequest): Promise<void> => {
    try {
      dispatch({ type: 'SET_AUTHENTICATING', payload: true });
      
      const response = await authService.login(credentials);
      const token = response.token;
      
      if (!isTokenValid(token)) {
        throw new Error('Invalid token received');
      }
      
      const user = getUserFromToken(token);
      if (!user) {
        throw new Error('Failed to parse user data from token');
      }
      
      tokenStorage.set(token);
      dispatch({ type: 'SET_USER', payload: { user, token } });
      
      toast.success(`Welcome back, ${user.handle}!`);
    } catch (error) {
      dispatch({ type: 'SET_AUTHENTICATING', payload: false });
      tokenStorage.remove();
      throw error;
    }
  }, []);

  const register = useCallback(async (userData: RegisterRequest): Promise<void> => {
    try {
      dispatch({ type: 'SET_AUTHENTICATING', payload: true });
      
      const response = await authService.register(userData);
      const token = response.token;
      
      if (!isTokenValid(token)) {
        throw new Error('Invalid token received');
      }
      
      const user = getUserFromToken(token);
      if (!user) {
        throw new Error('Failed to parse user data from token');
      }
      
      tokenStorage.set(token);
      dispatch({ type: 'SET_USER', payload: { user, token } });
      
      toast.success(`Welcome to the platform, ${user.handle}!`);
    } catch (error) {
      dispatch({ type: 'SET_AUTHENTICATING', payload: false });
      tokenStorage.remove();
      throw error;
    }
  }, []);

  const logout = useCallback(async (): Promise<void> => {
    tokenStorage.remove();
    dispatch({ type: 'CLEAR_USER' });
    toast.success('Logged out successfully');
  }, []);

  useEffect(() => {
    checkAuth();
  }, [checkAuth]);

  const value: AuthContextValue = useMemo(() => ({
    ...state,
    login,
    register,
    logout,
    checkAuth,
  }), [state, login, register, logout, checkAuth]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};