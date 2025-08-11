import { type ReactElement } from 'react'
import { render, type RenderOptions } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from '../../contexts/AuthContext'
import { ThemeProvider } from '../../contexts/ThemeContext'
import { AllTheProviders } from './providers'

interface RenderWithProvidersOptions extends Omit<RenderOptions, 'wrapper'> {
  initialAuth?: {
    isAuthenticated: boolean;
    user: any;
  };
  initialTheme?: 'light' | 'dark';
}

const createCustomProviders = (initialAuth?: any, initialTheme?: 'light' | 'dark') => {
  return ({ children }: { children: React.ReactNode }) => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
      },
    })

    const MockAuthProvider = ({ children }: { children: React.ReactNode }) => (
      <div data-testid="auth-context" data-auth={JSON.stringify(initialAuth)}>
        {children}
      </div>
    );

    const MockThemeProvider = ({ children }: { children: React.ReactNode }) => (
      <div data-testid="theme-context" data-theme={initialTheme}>
        {children}
      </div>
    );

    const AuthComponent = initialAuth ? MockAuthProvider : AuthProvider;
    const ThemeComponent = initialTheme ? MockThemeProvider : ThemeProvider;

    return (
      <BrowserRouter>
        <QueryClientProvider client={queryClient}>
          <ThemeComponent>
            <AuthComponent>
              {children}
            </AuthComponent>
          </ThemeComponent>
        </QueryClientProvider>
      </BrowserRouter>
    )
  }
}

export const renderWithProviders = (
  ui: ReactElement,
  options?: RenderWithProvidersOptions
) => {
  const { initialAuth, initialTheme, ...renderOptions } = options || {};
  
  if (initialAuth || initialTheme) {
    const CustomProviders = createCustomProviders(initialAuth, initialTheme);
    return render(ui, { wrapper: CustomProviders, ...renderOptions });
  }
  
  return render(ui, { wrapper: AllTheProviders, ...renderOptions });
};

const customRender = (
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>,
) => render(ui, { wrapper: AllTheProviders, ...options })

// eslint-disable-next-line react-refresh/only-export-components
export * from '@testing-library/react'
export { customRender as render }