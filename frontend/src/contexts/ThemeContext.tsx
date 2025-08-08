import type { ReactNode } from 'react';
import { useEffect, useState, useCallback, useMemo } from 'react';
import type { Theme, ThemeContextValue } from './theme-context';
import { ThemeContext } from './theme-context';

interface ThemeProviderProps {
  children: ReactNode;
}

const THEME_STORAGE_KEY = 'theme';

const getInitialTheme = (): Theme => {
  try {
    const savedTheme = localStorage.getItem(THEME_STORAGE_KEY) as Theme;
    if (savedTheme === 'light' || savedTheme === 'dark') {
      return savedTheme;
    }
  } catch (error) {
    console.error('Failed to read theme from localStorage:', error);
  }

  return 'dark';
};

const updateDocumentTheme = (theme: Theme) => {
  if (typeof document === 'undefined') return;
  
  const root = document.documentElement;
  
  if (theme === 'dark') {
    root.classList.add('dark');
  } else {
    root.classList.remove('dark');
  }

  root.setAttribute('data-theme', theme);
};

export const ThemeProvider = ({ children }: ThemeProviderProps) => {
  const [theme, setTheme] = useState<Theme>(() => {
    const initialTheme = getInitialTheme();
    updateDocumentTheme(initialTheme);
    return initialTheme;
  });

  const updateTheme = useCallback((newTheme: Theme) => {
    setTheme(newTheme);
    updateDocumentTheme(newTheme);
    
    try {
      localStorage.setItem(THEME_STORAGE_KEY, newTheme);
    } catch (error) {
      console.error('Failed to save theme to localStorage:', error);
    }
  }, []);

  const toggleTheme = useCallback(() => {
    const newTheme = theme === 'light' ? 'dark' : 'light';
    updateTheme(newTheme);
  }, [theme, updateTheme]);

  useEffect(() => {
    updateDocumentTheme(theme);
  }, [theme]);

  const value: ThemeContextValue = useMemo(() => ({
    theme,
    toggleTheme,
    setTheme: updateTheme,
  }), [theme, toggleTheme, updateTheme]);

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
};