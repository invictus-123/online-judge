import axios from 'axios';
import type { AxiosResponse } from 'axios';
import toast from 'react-hot-toast';
import type { ErrorResponse } from '../types/api';
import { tokenStorage } from './token-storage';

const API_BASE_URL = import.meta.env.VITE_HOST_NAME || 'http://localhost:8080';

const getApiUrl = () => {
  if (!API_BASE_URL) return 'http://localhost:8080';
  if (!API_BASE_URL.startsWith('http://') && !API_BASE_URL.startsWith('https://')) {
    return `https://${API_BASE_URL}`;
  }
  return API_BASE_URL;
};

export const apiClient = axios.create({
  baseURL: getApiUrl(),
  timeout: 10000,
  withCredentials: true,
});

apiClient.interceptors.request.use(
  (config) => {
    const token = tokenStorage.get();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    console.error('Request interceptor error:', error);
    throw error;
  }
);

apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    return response;
  },
  (error) => {
    if (error.response) {
      const errorData: ErrorResponse = error.response.data;
      
      if (error.response.status === 401 || error.response.status === 403) {
        const isLoginRequest = error.config?.url?.includes('/auth/login');
        
        if (!isLoginRequest) {
          tokenStorage.remove();
          toast.error('Session expired. Please log in again.');
          window.location.href = '/auth/login';
        } else {
          if (errorData.message) {
            error.message = errorData.message;
          }
        }
        throw error;
      }
      
      if (error.response.status === 429) {
        toast.error('Too many requests. Please try again later.');
        throw error;
      }
      
      if (error.response.status === 400 && errorData.details) {
        const isAuthRequest = error.config?.url?.includes('/auth/');
        if (!isAuthRequest) {
          const firstError = Object.values(errorData.details)[0];
          toast.error(firstError || errorData.message);
        }
        throw error;
      }
      
      if (error.response.status >= 500) {
        toast.error('Server error. Please try again later.');
        throw error;
      }
      
      const isAuthRequest = error.config?.url?.includes('/auth/');
      if (!isAuthRequest) {
        toast.error(errorData.message || 'An error occurred');
      } else {
        // For auth failures, enhance the error with backend message
        if (errorData.message) {
          error.message = errorData.message;
        }
      }
    } else if (error.request) {
      toast.error('Network error. Please check your connection.');
    } else {
      toast.error('An unexpected error occurred');
    }
    
    throw error;
  }
);

export default apiClient;