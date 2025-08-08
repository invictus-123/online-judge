import { describe, it, expect, vi, beforeEach } from 'vitest'
import { authService } from './auth-service'
import { apiClient } from './api-client'
import type { LoginRequest, RegisterRequest } from '../types/api'

vi.mock('./api-client')
const mockedApiClient = vi.mocked(apiClient)

describe('authService', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('login', () => {
    it('calls the login endpoint with correct data', async () => {
      const loginData: LoginRequest = {
        handle: 'testuser',
        password: 'password123'
      }
      
      const mockResponse = {
        data: {
          token: 'mock-jwt-token',
          user: {
            id: 1,
            handle: 'testuser',
            email: 'test@example.com',
            role: 'USER'
          }
        }
      }
      
      mockedApiClient.post.mockResolvedValue(mockResponse)
      
      const result = await authService.login(loginData)
      
      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v1/auth/login', loginData)
      expect(result).toEqual(mockResponse.data)
    })

    it('throws error when login fails', async () => {
      const loginData: LoginRequest = {
        handle: 'testuser',
        password: 'wrongpassword'
      }
      
      mockedApiClient.post.mockRejectedValue(new Error('Invalid credentials'))
      
      await expect(authService.login(loginData)).rejects.toThrow('Invalid credentials')
    })
  })

  describe('register', () => {
    it('calls the register endpoint with correct data', async () => {
      const registerData: RegisterRequest = {
        handle: 'newuser',
        email: 'newuser@example.com',
        firstName: 'John',
        lastName: 'Doe',
        password: 'Password123',
        userRole: 'USER'
      }
      
      const mockResponse = {
        data: {
          token: 'mock-jwt-token',
          user: {
            id: 2,
            handle: 'newuser',
            email: 'newuser@example.com',
            role: 'USER'
          }
        }
      }
      
      mockedApiClient.post.mockResolvedValue(mockResponse)
      
      const result = await authService.register(registerData)
      
      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v1/auth/register', registerData)
      expect(result).toEqual(mockResponse.data)
    })

    it('throws error when registration fails', async () => {
      const registerData: RegisterRequest = {
        handle: 'existinguser',
        email: 'existing@example.com',
        firstName: 'John',
        lastName: 'Doe',
        password: 'Password123',
        userRole: 'USER'
      }
      
      mockedApiClient.post.mockRejectedValue(new Error('User already exists'))
      
      await expect(authService.register(registerData)).rejects.toThrow('User already exists')
    })
  })

  describe('logout', () => {
    it('calls the logout endpoint', async () => {
      mockedApiClient.post.mockResolvedValue({ data: null })
      
      await authService.logout()
      
      expect(mockedApiClient.post).toHaveBeenCalledWith('/api/v1/auth/logout')
    })

    it('throws error when logout fails', async () => {
      mockedApiClient.post.mockRejectedValue(new Error('Logout failed'))
      
      await expect(authService.logout()).rejects.toThrow('Logout failed')
    })
  })
})