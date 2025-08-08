import type { AuthResponse, LoginRequest, RegisterRequest } from '../types/api';
import { apiClient } from './api-client';

export const authService = {
  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const response = await apiClient.post<AuthResponse>('/api/v1/auth/login', data);
    return response.data;
  },
  
  register: async (data: RegisterRequest): Promise<AuthResponse> => {
    const response = await apiClient.post<AuthResponse>('/api/v1/auth/register', data);
    return response.data;
  },
  
  logout: async (): Promise<void> => {
    await apiClient.post('/api/v1/auth/logout');
  }
};