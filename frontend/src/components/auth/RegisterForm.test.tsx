import { describe, it, expect, vi, beforeEach } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { render } from '../../test/utils/test-utils'
import { RegisterForm } from './RegisterForm'
import * as authHooks from '../../hooks/useAuth'
import { UserRole } from '../../types/api'

const mockRegister = vi.fn()
const mockUseAuth = vi.spyOn(authHooks, 'useAuth')

describe('RegisterForm', () => {
  const user = userEvent.setup()

  beforeEach(() => {
    vi.clearAllMocks()
    mockUseAuth.mockReturnValue({
      register: mockRegister,
      isAuthenticating: false,
      user: null,
      login: vi.fn(),
      logout: vi.fn(),
      isAuthenticated: false
    })
  })

  it('renders all form fields', () => {
    render(<RegisterForm />)
    
    expect(screen.getByRole('heading', { name: /create your account/i })).toBeInTheDocument()
    expect(screen.getByLabelText(/first name/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/last name/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/handle/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/role/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/confirm password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /create account/i })).toBeInTheDocument()
  })

  it('shows password visibility toggles', async () => {
    render(<RegisterForm />)
    
    const passwordInput = screen.getByLabelText(/^password$/i)
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)
    const toggleButtons = screen.getAllByRole('button', { name: '' })
    
    expect(passwordInput).toHaveAttribute('type', 'password')
    expect(confirmPasswordInput).toHaveAttribute('type', 'password')
    
    await user.click(toggleButtons[0])
    expect(passwordInput).toHaveAttribute('type', 'text')
    
    await user.click(toggleButtons[1])
    expect(confirmPasswordInput).toHaveAttribute('type', 'text')
  })

  it('validates all required fields', async () => {
    render(<RegisterForm />)
    
    const submitButton = screen.getByRole('button', { name: /create account/i })
    await user.click(submitButton)
    
    await waitFor(() => {
      expect(screen.getByText(/first name is required/i)).toBeInTheDocument()
      expect(screen.getByText(/last name is required/i)).toBeInTheDocument()
      expect(screen.getByText(/handle is required/i)).toBeInTheDocument()
      expect(screen.getByText(/email is required/i)).toBeInTheDocument()
      expect(screen.getByText(/password is required/i)).toBeInTheDocument()
    })
  })

  it('validates name fields format', async () => {
    render(<RegisterForm />)
    
    const firstNameInput = screen.getByLabelText(/first name/i)
    const lastNameInput = screen.getByLabelText(/last name/i)
    const submitButton = screen.getByRole('button', { name: /create account/i })
    
    await user.type(firstNameInput, 'John123')
    await user.type(lastNameInput, 'Doe@')
    await user.click(submitButton)
    
    await waitFor(() => {
      expect(screen.getByText(/first name must contain only letters/i)).toBeInTheDocument()
      expect(screen.getByText(/last name must contain only letters/i)).toBeInTheDocument()
    })
  })

  it('validates email format', async () => {
    render(<RegisterForm />)
    
    const emailInput = screen.getByLabelText(/email/i)
    const submitButton = screen.getByRole('button', { name: /create account/i })
    
    await user.type(emailInput, 'invalid-email')
    await user.click(submitButton)
    
    await waitFor(() => {
      expect(screen.getByText(/please provide a valid email address/i)).toBeInTheDocument()
    })
  })

  it('validates password strength', async () => {
    render(<RegisterForm />)
    
    const passwordInput = screen.getByLabelText(/^password$/i)
    const submitButton = screen.getByRole('button', { name: /create account/i })
    
    await user.type(passwordInput, 'weakpass')
    await user.click(submitButton)
    
    await waitFor(() => {
      expect(screen.getByText(/password must contain at least one uppercase letter, one lowercase letter, and one number/i)).toBeInTheDocument()
    })
  })

  it('shows password visibility toggles', async () => {
    render(<RegisterForm />)
    
    const passwordInput = screen.getByLabelText(/^password$/i)
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)
    
    expect(passwordInput).toHaveAttribute('type', 'password')
    expect(confirmPasswordInput).toHaveAttribute('type', 'password')
    
    await user.type(passwordInput, 'Password123')
    expect(passwordInput).toHaveDisplayValue('Password123')
  })

  it('validates password confirmation', async () => {
    render(<RegisterForm />)
    
    const passwordInput = screen.getByLabelText(/^password$/i)
    const confirmPasswordInput = screen.getByLabelText(/confirm password/i)
    const submitButton = screen.getByRole('button', { name: /create account/i })
    
    await user.type(passwordInput, 'Password123')
    await user.type(confirmPasswordInput, 'Different123')
    await user.click(submitButton)
    
    await waitFor(() => {
      expect(screen.getByText(/passwords don't match/i)).toBeInTheDocument()
    })
  })

  it('submits form with valid data', async () => {
    mockRegister.mockResolvedValue({})
    
    render(<RegisterForm />)
    
    await user.type(screen.getByLabelText(/first name/i), 'John')
    await user.type(screen.getByLabelText(/last name/i), 'Doe')
    await user.type(screen.getByLabelText(/handle/i), 'johndoe')
    await user.type(screen.getByLabelText(/email/i), 'john@example.com')
    await user.type(screen.getByLabelText(/^password$/i), 'Password123')
    await user.type(screen.getByLabelText(/confirm password/i), 'Password123')
    
    await user.click(screen.getByRole('button', { name: /create account/i }))
    
    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith({
        firstName: 'John',
        lastName: 'Doe',
        handle: 'johndoe',
        email: 'john@example.com',
        password: 'Password123',
        userRole: UserRole.USER
      })
    })
  })

  it('displays loading state during registration', () => {
    mockUseAuth.mockReturnValue({
      register: mockRegister,
      isAuthenticating: true,
      user: null,
      login: vi.fn(),
      logout: vi.fn(),
      isAuthenticated: false
    })
    
    render(<RegisterForm />)
    
    expect(screen.getByText(/creating account.../i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /creating account.../i })).toBeDisabled()
    expect(screen.getByLabelText(/first name/i)).toBeDisabled()
  })

  it('handles handle already exists error', async () => {
    mockRegister.mockRejectedValue(new Error('Handle already exists'))
    
    render(<RegisterForm />)
    
    await fillValidForm(user, screen)
    await user.click(screen.getByRole('button', { name: /create account/i }))
    
    await waitFor(() => {
      expect(screen.getByText(/this handle is already taken/i)).toBeInTheDocument()
    })
  })

  it('handles email already exists error', async () => {
    mockRegister.mockRejectedValue(new Error('Email already exists'))
    
    render(<RegisterForm />)
    
    await fillValidForm(user, screen)
    await user.click(screen.getByRole('button', { name: /create account/i }))
    
    await waitFor(() => {
      expect(screen.getByText(/this email is already registered/i)).toBeInTheDocument()
    })
  })

  it('handles validation errors', async () => {
    mockRegister.mockRejectedValue(new Error('Validation failed'))
    
    render(<RegisterForm />)
    
    await fillValidForm(user, screen)
    await user.click(screen.getByRole('button', { name: /create account/i }))
    
    await waitFor(() => {
      expect(screen.getByText(/please check your input and try again/i)).toBeInTheDocument()
    })
  })

  it('handles unexpected errors', async () => {
    mockRegister.mockRejectedValue('Network error')
    
    render(<RegisterForm />)
    
    await fillValidForm(user, screen)
    await user.click(screen.getByRole('button', { name: /create account/i }))
    
    await waitFor(() => {
      expect(screen.getByText(/an unexpected error occurred/i)).toBeInTheDocument()
    })
  })
})

async function fillValidForm(user: ReturnType<typeof userEvent.setup>, screen: typeof import('@testing-library/react').screen) {
  await user.type(screen.getByLabelText(/first name/i), 'John')
  await user.type(screen.getByLabelText(/last name/i), 'Doe')
  await user.type(screen.getByLabelText(/handle/i), 'johndoe')
  await user.type(screen.getByLabelText(/email/i), 'john@example.com')
  await user.type(screen.getByLabelText(/^password$/i), 'Password123')
  await user.type(screen.getByLabelText(/confirm password/i), 'Password123')
}