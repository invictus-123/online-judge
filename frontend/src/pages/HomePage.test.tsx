import { describe, it, expect, vi, beforeEach } from 'vitest'
import { screen } from '@testing-library/react'
import { render } from '../test/utils/test-utils'
import { HomePage } from './HomePage'
import * as authHooks from '../hooks/useAuth'

const mockUseAuth = vi.spyOn(authHooks, 'useAuth')

describe('HomePage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders homepage title and description', () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: false,
      user: null,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
      isAuthenticating: false,
      isInitializing: false
    })
    
    render(<HomePage />)
    
    expect(screen.getByRole('heading', { name: /online judge/i })).toBeInTheDocument()
    expect(screen.getByText(/master data structures & algorithms/i)).toBeInTheDocument()
  })

  it('shows welcome message for authenticated users', () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: true,
      user: { id: 1, handle: 'testuser', email: 'test@example.com', role: 'USER' },
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
      isAuthenticating: false,
      isInitializing: false
    })
    
    render(<HomePage />)
    
    expect(screen.getByText(/welcome back, testuser!/i)).toBeInTheDocument()
    expect(screen.getByText(/ready to solve some problems?/i)).toBeInTheDocument()
  })

  it('shows get started and sign in buttons for unauthenticated users', () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: false,
      user: null,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
      isAuthenticating: false,
      isInitializing: false
    })
    
    render(<HomePage />)
    
    expect(screen.getByRole('link', { name: /get started/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /sign in/i })).toBeInTheDocument()
  })

  it('renders feature cards', () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: false,
      user: null,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
      isAuthenticating: false,
      isInitializing: false
    })
    
    render(<HomePage />)
    
    expect(screen.getByText(/practice problems/i)).toBeInTheDocument()
    expect(screen.getByText(/track progress/i)).toBeInTheDocument()
    expect(screen.getByText(/multiple languages/i)).toBeInTheDocument()
  })

  it('renders platform statistics', () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: false,
      user: null,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
      isAuthenticating: false,
      isInitializing: false
    })
    
    render(<HomePage />)
    
    expect(screen.getByText(/platform statistics/i)).toBeInTheDocument()
    expect(screen.getByText(/500\+/)).toBeInTheDocument()
    expect(screen.getByText(/problems available/i)).toBeInTheDocument()
    expect(screen.getByText(/10k\+/)).toBeInTheDocument()
    expect(screen.getByText(/active users/i)).toBeInTheDocument()
  })

  it('shows different call-to-action for authenticated vs unauthenticated users in multiple languages card', () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: false,
      user: null,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
      isAuthenticating: false,
      isInitializing: false
    })
    
    const { rerender } = render(<HomePage />)
    expect(screen.getByRole('link', { name: /join now/i })).toBeInTheDocument()

    mockUseAuth.mockReturnValue({
      isAuthenticated: true,
      user: { id: 1, handle: 'testuser', email: 'test@example.com', role: 'USER' },
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
      isAuthenticating: false,
      isInitializing: false
    })
    
    rerender(<HomePage />)
    expect(screen.getByRole('link', { name: /start coding/i })).toBeInTheDocument()
  })

  it('shows sign up call-to-action for unauthenticated users', () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: false,
      user: null,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
      isAuthenticating: false,
      isInitializing: false
    })
    
    render(<HomePage />)
    
    expect(screen.getByText(/ready to start your journey?/i)).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /sign up now/i })).toBeInTheDocument()
  })

  it('hides sign up call-to-action for authenticated users', () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: true,
      user: { id: 1, handle: 'testuser', email: 'test@example.com', role: 'USER' },
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
      isAuthenticating: false,
      isInitializing: false
    })
    
    render(<HomePage />)
    
    expect(screen.queryByText(/ready to start your journey?/i)).not.toBeInTheDocument()
    expect(screen.queryByRole('link', { name: /sign up now/i })).not.toBeInTheDocument()
  })

  it('has correct navigation links', () => {
    mockUseAuth.mockReturnValue({
      isAuthenticated: false,
      user: null,
      login: vi.fn(),
      register: vi.fn(),
      logout: vi.fn(),
      isAuthenticating: false,
      isInitializing: false
    })
    
    render(<HomePage />)
    
    const browseProblemsLinks = screen.getAllByRole('link', { name: /browse problems/i })
    const viewSubmissionsLink = screen.getByRole('link', { name: /view submissions/i })
    
    expect(browseProblemsLinks).toHaveLength(2)
    expect(viewSubmissionsLink).toBeInTheDocument()
  })
})