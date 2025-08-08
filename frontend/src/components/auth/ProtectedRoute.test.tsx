import { describe, it, expect, vi, beforeEach } from 'vitest'
import { screen } from '@testing-library/react'
import { render } from '../../test/utils/test-utils'
import { ProtectedRoute } from './ProtectedRoute'
import * as authHooks from '../../hooks/useAuth'

const mockUseAuth = vi.spyOn(authHooks, 'useAuth')

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    Navigate: ({ to, state }: { to: string; state?: any }) => {
      mockNavigate(to, state)
      return <div data-testid="navigate">Navigate to: {to}</div>
    },
    useLocation: () => ({ pathname: '/protected', state: null })
  }
})

describe('ProtectedRoute', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading state when initializing', () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: false,
      isInitializing: true,
      user: null,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
      isAuthenticating: false
    })
    
    render(
      <ProtectedRoute>
        <div>Protected content</div>
      </ProtectedRoute>
    )
    
    expect(screen.getByText(/loading.../i)).toBeInTheDocument()
    expect(screen.queryByText('Protected content')).not.toBeInTheDocument()
  })

  it('redirects to login when not authenticated and auth is required', () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: false,
      isInitializing: false,
      user: null,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
      isAuthenticating: false
    })
    
    render(
      <ProtectedRoute requireAuth={true}>
        <div>Protected content</div>
      </ProtectedRoute>
    )
    
    expect(mockNavigate).toHaveBeenCalledWith('/auth/login', { from: { pathname: '/protected', state: null } })
    expect(screen.queryByText('Protected content')).not.toBeInTheDocument()
  })

  it('renders children when authenticated and auth is required', () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: true,
      isInitializing: false,
      user: { id: 1, handle: 'testuser', email: 'test@example.com', role: 'USER' },
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
      isAuthenticating: false
    })
    
    render(
      <ProtectedRoute requireAuth={true}>
        <div>Protected content</div>
      </ProtectedRoute>
    )
    
    expect(screen.getByText('Protected content')).toBeInTheDocument()
  })

  it('renders children when not authenticated and auth is not required', () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: false,
      isInitializing: false,
      user: null,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
      isAuthenticating: false
    })
    
    render(
      <ProtectedRoute requireAuth={false}>
        <div>Public content</div>
      </ProtectedRoute>
    )
    
    expect(screen.getByText('Public content')).toBeInTheDocument()
  })

  it('redirects authenticated users away from public routes', () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: true,
      isInitializing: false,
      user: { id: 1, handle: 'testuser', email: 'test@example.com', role: 'USER' },
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
      isAuthenticating: false
    })
    
    render(
      <ProtectedRoute requireAuth={false}>
        <div>Public content</div>
      </ProtectedRoute>
    )
    
    expect(mockNavigate).toHaveBeenCalledWith('/', undefined)
    expect(screen.queryByText('Public content')).not.toBeInTheDocument()
  })

  it('uses default requireAuth=true when not specified', () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: false,
      isInitializing: false,
      user: null,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
      isAuthenticating: false
    })
    
    render(
      <ProtectedRoute>
        <div>Protected content</div>
      </ProtectedRoute>
    )
    
    expect(mockNavigate).toHaveBeenCalledWith('/auth/login', { from: { pathname: '/protected', state: null } })
  })
})