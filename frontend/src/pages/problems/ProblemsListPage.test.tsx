import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter } from 'react-router-dom';
import { ProblemsListPage } from './ProblemsListPage';
import { AuthProvider } from '../../contexts/AuthContext';
import { problemsService } from '../../services/problems-service';
import type { ListProblemsResponse, ProblemSummaryUi } from '../../types/api';

vi.mock('../../services/problems-service');

vi.mock('../../hooks/useAuth', () => ({
  useAuth: () => ({
    isAuthenticated: true,
    user: { id: '1', handle: 'testuser' }
  })
}));

const mockProblems: ProblemSummaryUi[] = [
  {
    id: 1,
    title: 'Two Sum',
    difficulty: 'EASY',
    tags: ['ARRAY', 'GREEDY'],
    solvedStatus: 'SOLVED'
  },
  {
    id: 2,
    title: 'Add Two Numbers',
    difficulty: 'MEDIUM',
    tags: ['STRING', 'DP'],
    solvedStatus: 'ATTEMPTED'
  },
  {
    id: 3,
    title: 'Median of Two Sorted Arrays',
    difficulty: 'HARD',
    tags: ['TREE', 'GRAPH'],
    solvedStatus: 'UNATTEMPTED'
  }
];

const mockResponse: ListProblemsResponse = {
  problems: mockProblems
};

function TestWrapper({ children }: { children: React.ReactNode }) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AuthProvider>
          {children}
        </AuthProvider>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

describe('ProblemsListPage', () => {
  it('renders page title and description', () => {
    vi.mocked(problemsService.list).mockResolvedValue(mockResponse);
    
    render(
      <TestWrapper>
        <ProblemsListPage />
      </TestWrapper>
    );

    expect(screen.getByText('Problems')).toBeInTheDocument();
    expect(screen.getByText('Solve coding challenges and improve your skills')).toBeInTheDocument();
  });

  it('displays loading state', () => {
    vi.mocked(problemsService.list).mockImplementation(() => new Promise(() => {}));
    
    render(
      <TestWrapper>
        <ProblemsListPage />
      </TestWrapper>
    );

    const loadingElements = document.querySelectorAll('.animate-pulse');
    expect(loadingElements.length).toBeGreaterThan(0);
  });

  it('renders problem cards with correct information', async () => {
    vi.mocked(problemsService.list).mockResolvedValue(mockResponse);
    
    render(
      <TestWrapper>
        <ProblemsListPage />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('1. Two Sum')).toBeInTheDocument();
      expect(screen.getByText('2. Add Two Numbers')).toBeInTheDocument();
      expect(screen.getByText('3. Median of Two Sorted Arrays')).toBeInTheDocument();
    });

    expect(screen.getByText('EASY')).toBeInTheDocument();
    expect(screen.getByText('MEDIUM')).toBeInTheDocument();
    expect(screen.getByText('HARD')).toBeInTheDocument();

    const problemCards = screen.getAllByRole('link');
    expect(problemCards.length).toBeGreaterThan(0);
    
    const tagElements = document.querySelectorAll('.px-2.py-1.text-xs.rounded-md');
    expect(tagElements.length).toBeGreaterThan(0);
  });

  it('displays solved status icons for authenticated users', async () => {
    vi.mocked(problemsService.list).mockResolvedValue(mockResponse);
    
    render(
      <TestWrapper>
        <ProblemsListPage />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('✅')).toBeInTheDocument();
      expect(screen.getByText('❌')).toBeInTheDocument();
    });
  });

  it('renders filter controls', async () => {
    vi.mocked(problemsService.list).mockResolvedValue(mockResponse);
    
    render(
      <TestWrapper>
        <ProblemsListPage />
      </TestWrapper>
    );

    await waitFor(() => {
      const selects = screen.getAllByRole('combobox');
      expect(selects.length).toBeGreaterThanOrEqual(3);
    });
  });

  it('renders pagination controls', async () => {
    vi.mocked(problemsService.list).mockResolvedValue(mockResponse);
    
    render(
      <TestWrapper>
        <ProblemsListPage />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('Previous')).toBeInTheDocument();
      expect(screen.getByText('Next')).toBeInTheDocument();
      expect(screen.getByText('Page 1')).toBeInTheDocument();
    });
  });

  it('displays error state when API call fails', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    vi.mocked(problemsService.list).mockRejectedValue(new Error('API Error'));
    
    render(
      <TestWrapper>
        <ProblemsListPage />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('Error Loading Problems')).toBeInTheDocument();
      expect(screen.getByText('Try Again')).toBeInTheDocument();
    });
    
    consoleSpy.mockRestore();
  });

  it('displays no problems message when list is empty', async () => {
    vi.mocked(problemsService.list).mockResolvedValue({ problems: [] });
    
    render(
      <TestWrapper>
        <ProblemsListPage />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('No problems found')).toBeInTheDocument();
    });
  });
});