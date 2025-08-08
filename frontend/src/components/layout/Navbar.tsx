import { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { 
  Menu, 
  X, 
  Sun, 
  Moon, 
  Code2, 
  User, 
  LogOut, 
  Settings,
  FileText,
  BarChart3
} from 'lucide-react';
import { useAuth, useTheme } from '../../hooks';
import { Button } from '../ui';

interface NavLinkProps {
  to: string;
  children: React.ReactNode;
  mobile?: boolean;
  onClick?: () => void;
}

const NavLink = ({ to, children, mobile = false, onClick }: NavLinkProps) => {
  const location = useLocation();
  const isActive = location.pathname === to;
  
  const baseClasses = mobile
    ? 'flex items-center px-3 py-2 text-base font-medium rounded-md transition-colors'
    : 'inline-flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors';
    
  const activeClasses = isActive
    ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-200 border border-blue-200 dark:border-blue-700'
    : 'text-gray-700 hover:text-gray-900 hover:bg-gray-100 dark:text-gray-300 dark:hover:text-slate-100 dark:hover:bg-slate-800';

  return (
    <Link
      to={to}
      className={`${baseClasses} ${activeClasses}`}
      onClick={onClick}
    >
      {children}
    </Link>
  );
};

export const Navbar = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);
  const { user, isAuthenticated, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();

  const handleLogout = async () => {
    await logout();
    setIsUserMenuOpen(false);
    setIsMobileMenuOpen(false);
  };

  const closeMobileMenu = () => {
    setIsMobileMenuOpen(false);
  };

  const toggleUserMenu = () => {
    setIsUserMenuOpen(!isUserMenuOpen);
  };

  return (
    <nav className="shadow-sm dark:shadow-gray-700/10">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 justify-between">
          <div className="flex items-center space-x-8">
            <Link 
              to="/" 
              className="flex items-center space-x-2 text-xl font-bold text-gray-900 hover:text-blue-600 dark:text-white dark:hover:text-blue-400"
            >
              <Code2 className="h-6 w-6" />
              <span>Online Judge</span>
            </Link>
            
            <div className="hidden md:flex md:items-center md:space-x-4">
              <NavLink to="/problems">
                <FileText className="mr-2 h-4 w-4" />
                Problems
              </NavLink>
              <NavLink to="/submissions">
                <BarChart3 className="mr-2 h-4 w-4" />
                Submissions
              </NavLink>
            </div>
          </div>

          <div className="hidden md:flex md:items-center md:space-x-4">
            <Button
              variant="ghost"
              size="sm"
              onClick={toggleTheme}
              className="ml-4"
              aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
              title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
            >
              {theme === 'dark' ? (
                <Sun className="h-4 w-4" />
              ) : (
                <Moon className="h-4 w-4" />
              )}
            </Button>

            {isAuthenticated ? (
              <div className="relative ml-4">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={toggleUserMenu}
                  className="flex items-center space-x-2"
                  aria-label={isUserMenuOpen ? 'Close user menu' : 'Open user menu'}
                  aria-expanded={isUserMenuOpen}
                  aria-haspopup="menu"
                >
                  <User className="h-4 w-4" />
                  <span>{user?.handle}</span>
                </Button>

                {isUserMenuOpen && (
                  <>
                    <button
                      type="button"
                      className="fixed inset-0 z-40 cursor-default"
                      onClick={() => setIsUserMenuOpen(false)}
                      onKeyDown={(e) => {
                        if (e.key === 'Escape') {
                          setIsUserMenuOpen(false);
                        }
                      }}
                      aria-label="Close user menu"
                    />

                    <div className="absolute right-0 z-50 mt-2 w-48 rounded-md bg-white py-1 shadow-lg ring-1 ring-black ring-opacity-5 dark:bg-gray-800 dark:ring-gray-700">
                      <div className="px-4 py-2 text-sm text-gray-500 dark:text-gray-400">
                        Signed in as <span className="font-medium text-gray-900 dark:text-white">{user?.handle}</span>
                      </div>
                      <hr className="border-gray-200 dark:border-gray-700" />
                      
                      <Link
                        to="/profile"
                        className="flex items-center px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
                        onClick={() => setIsUserMenuOpen(false)}
                      >
                        <Settings className="mr-3 h-4 w-4" />
                        Profile Settings
                      </Link>
                      
                      <button
                        onClick={handleLogout}
                        className="flex w-full items-center px-4 py-2 text-sm text-red-700 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
                      >
                        <LogOut className="mr-3 h-4 w-4" />
                        Sign out
                      </button>
                    </div>
                  </>
                )}
              </div>
            ) : (
              <div className="ml-4 space-x-2">
                <Link to="/auth/login">
                  <Button variant="ghost" size="sm">Sign in</Button>
                </Link>
                <Link to="/auth/register">
                  <Button size="sm">Sign up</Button>
                </Link>
              </div>
            )}
          </div>

          <div className="flex items-center space-x-2 md:hidden">
            <Button
              variant="ghost"
              size="sm"
              onClick={toggleTheme}
              aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
              title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
            >
              {theme === 'dark' ? (
                <Sun className="h-4 w-4" />
              ) : (
                <Moon className="h-4 w-4" />
              )}
            </Button>

            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
              aria-label={isMobileMenuOpen ? 'Close mobile menu' : 'Open mobile menu'}
              aria-expanded={isMobileMenuOpen}
            >
              {isMobileMenuOpen ? (
                <X className="h-6 w-6" />
              ) : (
                <Menu className="h-6 w-6" />
              )}
            </Button>
          </div>
        </div>
      </div>

      {isMobileMenuOpen && (
        <>
          <button
            type="button"
            className="fixed inset-0 z-20 bg-black bg-opacity-25 cursor-default md:hidden"
            onClick={closeMobileMenu}
            onKeyDown={(e) => {
              if (e.key === 'Escape') {
                closeMobileMenu();
              }
            }}
            aria-label="Close mobile menu"
          />
          
          <div className="absolute left-0 right-0 z-30 bg-white shadow-lg dark:bg-gray-800 md:hidden">
            <div className="space-y-1 px-4 pb-3 pt-2">
              <NavLink to="/problems" mobile onClick={closeMobileMenu}>
                <FileText className="mr-3 h-5 w-5" />
                Problems
              </NavLink>
              <NavLink to="/submissions" mobile onClick={closeMobileMenu}>
                <BarChart3 className="mr-3 h-5 w-5" />
                Submissions
              </NavLink>

              {isAuthenticated ? (
                <>
                  <hr className="my-3 border-gray-200 dark:border-gray-700" />
                  <div className="px-3 py-2 text-sm text-gray-500 dark:text-gray-400">
                    Signed in as <span className="font-medium text-gray-900 dark:text-white">{user?.handle}</span>
                  </div>
                  
                  <NavLink to="/profile" mobile onClick={closeMobileMenu}>
                    <Settings className="mr-3 h-5 w-5" />
                    Profile Settings
                  </NavLink>
                  
                  <button
                    onClick={handleLogout}
                    className="flex w-full items-center px-3 py-2 text-base font-medium text-red-700 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20 rounded-md transition-colors"
                  >
                    <LogOut className="mr-3 h-5 w-5" />
                    Sign out
                  </button>
                </>
              ) : (
                <>
                  <hr className="my-3 border-gray-200 dark:border-gray-700" />
                  <NavLink to="/auth/login" mobile onClick={closeMobileMenu}>
                    Sign in
                  </NavLink>
                  <NavLink to="/auth/register" mobile onClick={closeMobileMenu}>
                    Sign up
                  </NavLink>
                </>
              )}
            </div>
          </div>
        </>
      )}
    </nav>
  );
};