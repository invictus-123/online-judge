import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useNavigate } from 'react-router-dom';
import { Eye, EyeOff } from 'lucide-react';
import { useAuth } from '../../hooks';
import { Button, Input, Select, Card, CardHeader, CardTitle, CardContent } from '../ui';
import { UserRole } from '../../types/api';
import type { RegisterRequest } from '../../types/api';

const registerSchema = z.object({
  handle: z
    .string()
    .min(1, 'Handle is required')
    .min(3, 'Handle must be at least 3 characters')
    .max(20, 'Handle must not exceed 20 characters')
    .regex(/^[a-zA-Z0-9_-]+$/, 'Handle can only contain letters, numbers, underscores, and hyphens'),
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Please provide a valid email address'),
  firstName: z
    .string()
    .min(1, 'First name is required')
    .regex(/^[a-zA-Z]*$/, 'First name must contain only letters')
    .max(50, 'First name must not exceed 50 characters'),
  lastName: z
    .string()
    .min(1, 'Last name is required')
    .regex(/^[a-zA-Z]*$/, 'Last name must contain only letters')
    .max(50, 'Last name must not exceed 50 characters'),
  password: z
    .string()
    .min(1, 'Password is required')
    .min(8, 'Password must be at least 8 characters long')
    .regex(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/, 'Password must contain at least one uppercase letter, one lowercase letter, and one number'),
  confirmPassword: z
    .string()
    .min(1, 'Please confirm your password'),
  userRole: z
    .enum([UserRole.USER, UserRole.ADMIN] as const)
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

type RegisterFormData = z.infer<typeof registerSchema>;

const roleOptions = [
  { value: UserRole.USER, label: 'User' },
  { value: UserRole.ADMIN, label: 'Admin' }
];

export const RegisterForm = () => {
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [registerError, setRegisterError] = useState<string | null>(null);
  const { register: registerUser, isAuthenticating } = useAuth();
  const navigate = useNavigate();

  const {
    register,
    handleSubmit,
    formState: { errors },
    setError,
    watch
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    mode: 'onChange',
    defaultValues: {
      handle: '',
      email: '',
      firstName: '',
      lastName: '',
      password: '',
      confirmPassword: '',
      userRole: UserRole.USER
    }
  });

  const password = watch('password');

  const onSubmit = async (data: RegisterFormData) => {
    try {
      setRegisterError(null);
      const registerRequest: RegisterRequest = {
        handle: data.handle,
        email: data.email,
        firstName: data.firstName,
        lastName: data.lastName,
        password: data.password,
        userRole: data.userRole
      };

      await registerUser(registerRequest);
      navigate('/', { replace: true });
    } catch (error) {
      if (error instanceof Error) {
        const message = error.message.toLowerCase();
        if (message.includes('handle') && message.includes('exists')) {
          setError('handle', { message: 'This handle is already taken' });
        } else if (message.includes('email') && message.includes('exists')) {
          setError('email', { message: 'This email is already registered' });
        } else if (message.includes('validation')) {
          setRegisterError('Please check your input and try again');
        } else {
          setRegisterError(error.message || 'Registration failed. Please try again.');
        }
      } else {
        setRegisterError('An unexpected error occurred');
      }
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12 dark:bg-gray-900 sm:px-6 lg:px-8">
      <div className="w-full max-w-md space-y-8">
        <div className="text-center">
          <h2 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
            Create your account
          </h2>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Register</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <Input
                  label="First Name"
                  type="text"
                  autoComplete="given-name"
                  placeholder="Enter your first name"
                  error={errors.firstName?.message}
                  disabled={isAuthenticating}
                  {...register('firstName')}
                />

                <Input
                  label="Last Name"
                  type="text"
                  autoComplete="family-name"
                  placeholder="Enter your last name"
                  error={errors.lastName?.message}
                  disabled={isAuthenticating}
                  {...register('lastName')}
                />
              </div>

              <Input
                label="Handle"
                type="text"
                autoComplete="username"
                placeholder="3-20 chars: letters, numbers, _, -"
                error={errors.handle?.message}
                disabled={isAuthenticating}
                {...register('handle')}
              />

              <Input
                label="Email"
                type="email"
                autoComplete="email"
                placeholder="Enter your email address"
                error={errors.email?.message}
                disabled={isAuthenticating}
                {...register('email')}
              />

              <Select
                label="Role"
                options={roleOptions}
                error={errors.userRole?.message}
                disabled={isAuthenticating}
                {...register('userRole')}
              />

              <div className="relative">
                <Input
                  label="Password"
                  type={showPassword ? 'text' : 'password'}
                  autoComplete="new-password"
                  placeholder="8+ chars: uppercase, lowercase, number"
                  error={errors.password?.message}
                  disabled={isAuthenticating}
                  {...register('password')}
                  className="pr-10"
                />
                <button
                  type="button"
                  className="absolute right-3 top-[38px] text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300"
                  onClick={() => setShowPassword(!showPassword)}
                  disabled={isAuthenticating}
                >
                  {showPassword ? (
                    <EyeOff className="h-5 w-5" />
                  ) : (
                    <Eye className="h-5 w-5" />
                  )}
                </button>
              </div>

              <div className="relative">
                <Input
                  label="Confirm Password"
                  type={showConfirmPassword ? 'text' : 'password'}
                  autoComplete="new-password"
                  placeholder="Confirm your password"
                  error={errors.confirmPassword?.message}
                  disabled={isAuthenticating}
                  {...register('confirmPassword')}
                  className="pr-10"
                />
                <button
                  type="button"
                  className="absolute right-3 top-[38px] text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300"
                  onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                  disabled={isAuthenticating}
                >
                  {showConfirmPassword ? (
                    <EyeOff className="h-5 w-5" />
                  ) : (
                    <Eye className="h-5 w-5" />
                  )}
                </button>
              </div>

              {/* Compact Password strength indicator */}
              {password && (
                <div className="flex items-center gap-4 text-xs">
                  <div className={`flex items-center gap-1 ${password.length >= 8 ? 'text-green-600 dark:text-green-400' : 'text-gray-400'}`}>
                    {password.length >= 8 ? '✓' : '○'} 8+
                  </div>
                  <div className={`flex items-center gap-1 ${/[a-z]/.test(password) ? 'text-green-600 dark:text-green-400' : 'text-gray-400'}`}>
                    {/[a-z]/.test(password) ? '✓' : '○'} abc
                  </div>
                  <div className={`flex items-center gap-1 ${/[A-Z]/.test(password) ? 'text-green-600 dark:text-green-400' : 'text-gray-400'}`}>
                    {/[A-Z]/.test(password) ? '✓' : '○'} ABC
                  </div>
                  <div className={`flex items-center gap-1 ${/\d/.test(password) ? 'text-green-600 dark:text-green-400' : 'text-gray-400'}`}>
                    {/\d/.test(password) ? '✓' : '○'} 123
                  </div>
                </div>
              )}

              {(errors.root || registerError) && (
                <div className="rounded-md bg-red-50 p-4 dark:bg-red-900/20">
                  <div className="text-sm text-red-700 dark:text-red-400">
                    {registerError || errors.root?.message}
                  </div>
                </div>
              )}

              <Button
                type="submit"
                className="w-full"
                isLoading={isAuthenticating}
                disabled={isAuthenticating}
              >
                {isAuthenticating ? 'Creating account...' : 'Create account'}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};