import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from 'react-hot-toast';
import { AuthProvider } from './contexts/AuthContext';
import { ThemeProvider } from './contexts/ThemeContext';
import { Layout, AuthLayout } from './components/layout';
import { ProtectedRoute } from './components/auth/ProtectedRoute';
import { HomePage } from './pages/HomePage';
import { LoginPage } from './pages/auth/LoginPage';
import { RegisterPage } from './pages/auth/RegisterPage';
import { ProblemsListPage, ProblemDetailPage } from './pages/problems';
import { SubmissionsListPage } from './pages/submissions';
import './index.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <AuthProvider>
          <Router>
            <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
              <Routes>
                <Route path="/auth/login" element={
                  <ProtectedRoute requireAuth={false}>
                    <AuthLayout>
                      <LoginPage />
                    </AuthLayout>
                  </ProtectedRoute>
                } />
                <Route path="/auth/register" element={
                  <ProtectedRoute requireAuth={false}>
                    <AuthLayout>
                      <RegisterPage />
                    </AuthLayout>
                  </ProtectedRoute>
                } />

                <Route path="/" element={<Layout />}>
                  <Route index element={<HomePage />} />
                  
                  <Route path="/problems" element={<ProblemsListPage />} />
                  <Route path="/problems/:id" element={<ProblemDetailPage />} />
                  
                  <Route path="/submissions" element={<SubmissionsListPage />} />
                  
                  <Route path="/profile" element={
                    <ProtectedRoute>
                      <div className="flex min-h-[50vh] items-center justify-center">
                        <div className="text-center">
                          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Profile Page</h2>
                          <p className="text-gray-600 dark:text-gray-400">Coming soon...</p>
                        </div>
                      </div>
                    </ProtectedRoute>
                  } />
                </Route>

                <Route path="*" element={<Navigate to="/" replace />} />
              </Routes>
            </div>
          </Router>

          <Toaster
            position="top-right"
            toastOptions={{
              duration: 4000,
              style: {
                background: 'var(--toast-bg)',
                color: 'var(--toast-color)',
                border: '1px solid var(--toast-border)',
              },
            }}
          />
        </AuthProvider>
      </ThemeProvider>
    </QueryClientProvider>
  );
}

export default App;