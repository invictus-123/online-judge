import { describe, it, expect, vi, beforeEach } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { render } from '../../test/utils/test-utils'
import { LoginForm } from './LoginForm'
import * as authHooks from '../../hooks/useAuth'

const mockLogin = vi.fn()
const mockUseAuth = vi.spyOn(authHooks, 'useAuth')

describe('LoginForm', () => {
  const user = userEvent.setup()

  beforeEach(() => {
    vi.clearAllMocks()
    mockUseAuth.mockReturnValue({
      login: mockLogin,
      isAuthenticating: false,
      user: null,
      logout: vi.fn(),
      isAuthenticated: false
    })
  })

  it('renders login form elements', () => {
    render(<LoginForm />)
    
    expect(screen.getByRole('heading', { name: /sign in to your account/i })).toBeInTheDocument()
    expect(screen.getByLabelText(/handle/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('shows password visibility toggle', async () => {
    render(<LoginForm />)
    
    const passwordInput = screen.getByLabelText(/password/i)
    const toggleButton = screen.getByRole('button', { name: '' })
    
    expect(passwordInput).toHaveAttribute('type', 'password')
    
    await user.click(toggleButton)
    expect(passwordInput).toHaveAttribute('type', 'text')
    
    await user.click(toggleButton)
    expect(passwordInput).toHaveAttribute('type', 'password')
  })

  it('validates required fields', async () => {
    render(<LoginForm />)
    
    const submitButton = screen.getByRole('button', { name: /sign in/i })
    await user.click(submitButton)
    
    await waitFor(() => {
      expect(screen.getByText(/handle is required/i)).toBeInTheDocument()
      expect(screen.getByText(/password is required/i)).toBeInTheDocument()
    })
  })

  it('validates handle format and length', async () => {
    render(<LoginForm />)
    
    const handleInput = screen.getByLabelText(/handle/i)
    const submitButton = screen.getByRole('button', { name: /sign in/i })
    
    await user.type(handleInput, 'ab')
    await user.click(submitButton)
    
    await waitFor(() => {
      expect(screen.getByText(/handle must be at least 3 characters/i)).toBeInTheDocument()
    })
    
    await user.clear(handleInput)
    await user.type(handleInput, 'a'.repeat(21))
    await user.click(submitButton)
    
    await waitFor(() => {
      expect(screen.getByText(/handle must not exceed 20 characters/i)).toBeInTheDocument()
    })
    
    await user.clear(handleInput)
    await user.type(handleInput, 'user@domain')
    await user.click(submitButton)
    
    await waitFor(() => {
      expect(screen.getByText(/handle can only contain letters, numbers, underscores, and hyphens/i)).toBeInTheDocument()
    })
  })

  it('validates password length', async () => {
    render(<LoginForm />)
    
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /sign in/i })
    
    await user.type(passwordInput, '1234567')
    await user.click(submitButton)
    
    await waitFor(() => {
      expect(screen.getByText(/password must be at least 8 characters long/i)).toBeInTheDocument()
    })
  })

  it('submits form with valid data', async () => {
    mockLogin.mockResolvedValue({})
    
    render(<LoginForm />)
    
    const handleInput = screen.getByLabelText(/handle/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /sign in/i })
    
    await user.type(handleInput, 'testuser')
    await user.type(passwordInput, 'password123')
    await user.click(submitButton)
    
    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({
        handle: 'testuser',
        password: 'password123'
      })
    })
  })

  it('displays loading state during authentication', () => {
    mockUseAuth.mockReturnValue({
      login: mockLogin,
      isAuthenticating: true,
      user: null,
      logout: vi.fn(),
      isAuthenticated: false
    })
    
    render(<LoginForm />)
    
    expect(screen.getByText(/signing in.../i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /signing in.../i })).toBeDisabled()
    expect(screen.getByLabelText(/handle/i)).toBeDisabled()
    expect(screen.getByLabelText(/password/i)).toBeDisabled()
  })

  it('handles login errors', async () => {
    const errorMessage = 'Invalid username or password'
    mockLogin.mockRejectedValue(new Error(errorMessage))
    
    render(<LoginForm />)
    
    const handleInput = screen.getByLabelText(/handle/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /sign in/i })
    
    await user.type(handleInput, 'testuser')
    await user.type(passwordInput, 'wrongpassword')
    await user.click(submitButton)
    
    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument()
    })
  })

  it('handles user not found error', async () => {
    mockLogin.mockRejectedValue(new Error('User not found'))
    
    render(<LoginForm />)
    
    const handleInput = screen.getByLabelText(/handle/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /sign in/i })
    
    await user.type(handleInput, 'nonexistentuser')
    await user.type(passwordInput, 'password123')
    await user.click(submitButton)
    
    await waitFor(() => {
      expect(screen.getByText(/user not found/i)).toBeInTheDocument()
    })
  })

  it('handles unexpected errors', async () => {
    mockLogin.mockRejectedValue('Network error')
    
    render(<LoginForm />)
    
    const handleInput = screen.getByLabelText(/handle/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /sign in/i })
    
    await user.type(handleInput, 'testuser')
    await user.type(passwordInput, 'password123')
    await user.click(submitButton)
    
    await waitFor(() => {
      expect(screen.getByText(/an unexpected error occurred/i)).toBeInTheDocument()
    })
  })
})