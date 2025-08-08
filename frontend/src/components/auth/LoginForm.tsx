import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useNavigate, useLocation } from 'react-router-dom';
import { Eye, EyeOff } from 'lucide-react';
import { useAuth } from '../../hooks';
import { Button, Input, Card, CardHeader, CardTitle, CardContent } from '../ui';
import type { LoginRequest } from '../../types/api';

const loginSchema = z.object({
  handle: z
    .string()
    .min(1, 'Handle is required')
    .min(3, 'Handle must be at least 3 characters')
    .max(20, 'Handle must not exceed 20 characters')
    .regex(/^[a-zA-Z0-9_-]+$/, 'Handle can only contain letters, numbers, underscores, and hyphens'),
  password: z
    .string()
    .min(1, 'Password is required')
    .min(8, 'Password must be at least 8 characters long')
});

type LoginFormData = z.infer<typeof loginSchema>;

export const LoginForm = () => {
  const [showPassword, setShowPassword] = useState(false);
  const [loginError, setLoginError] = useState<string | null>(null);
  const { login, isAuthenticating } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const from = location.state?.from?.pathname || '/';

  const {
    register,
    handleSubmit,
    formState: { errors },
    setError
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      handle: '',
      password: ''
    }
  });

  const onSubmit = async (data: LoginFormData) => {
    try {
      setLoginError(null);
      const loginRequest: LoginRequest = {
        handle: data.handle,
        password: data.password
      };

      await login(loginRequest);
      navigate(from, { replace: true });
    } catch (error) {
      if (error instanceof Error) {
        const message = error.message.toLowerCase();
        if (message.includes('invalid') || message.includes('wrong') || message.includes('incorrect') || message.includes('bad_credentials')) {
          setLoginError(error.message || 'Invalid username or password');
        } else if (message.includes('not found') || message.includes('does not exist')) {
          setError('handle', { message: 'User not found' });
        } else {
          setLoginError(error.message || 'Login failed. Please try again.');
        }
      } else {
        setLoginError('An unexpected error occurred');
      }
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12 dark:bg-gray-900 sm:px-6 lg:px-8">
      <div className="w-full max-w-md space-y-8">
        <div className="text-center">
          <h2 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
            Sign in to your account
          </h2>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Login</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={(e) => {
              e.preventDefault();
              handleSubmit(onSubmit)(e);
            }} className="space-y-6">
              <Input
                label="Handle"
                type="text"
                autoComplete="username"
                placeholder="Enter your handle"
                error={errors.handle?.message}
                disabled={isAuthenticating}
                {...register('handle')}
              />

              <div className="relative">
                <Input
                  label="Password"
                  type={showPassword ? 'text' : 'password'}
                  autoComplete="current-password"
                  placeholder="Enter your password"
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

              {loginError && (
                <div className="rounded-md bg-red-50 p-4 dark:bg-red-900/20">
                  <div className="text-sm text-red-700 dark:text-red-400">
                    {loginError}
                  </div>
                </div>
              )}

              <Button
                type="submit"
                className="w-full"
                isLoading={isAuthenticating}
                disabled={isAuthenticating}
              >
                {isAuthenticating ? 'Signing in...' : 'Sign in'}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};