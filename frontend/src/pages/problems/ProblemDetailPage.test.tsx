import { describe, it, expect, beforeEach, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { renderWithProviders, server } from '../../test/utils';
import { ProblemDetailPage } from './ProblemDetailPage';
import { SolvedStatus } from '../../types/api';
import { useAuth } from '../../hooks/useAuth';
import { useTheme } from '../../hooks/useTheme';
import { useDocumentTitle } from '../../hooks/useDocumentTitle';

// Mock the hooks
vi.mock('../../hooks/useAuth');
vi.mock('../../hooks/useTheme');
vi.mock('../../hooks/useDocumentTitle');

// Mock react-router-dom useParams
const mockUseParams = vi.fn();
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useParams: () => mockUseParams(),
    Navigate: ({ to, replace }: { to: string, replace?: boolean }) => <div data-testid="navigate" data-to={to} data-replace={replace} />
  };
});

// Mock monaco editor
vi.mock('@monaco-editor/react', () => ({
  Editor: ({ value, onChange, language, theme }: {
    value?: string;
    onChange?: (value: string | undefined) => void;
    language?: string;
    theme?: string;
  }) => (
    <textarea
      data-testid="monaco-editor"
      data-language={language}
      data-theme={theme}
      value={value}
      onChange={(e) => onChange?.(e.target.value)}
      className="w-full h-full"
    />
  )
}));

const mockProblemDetails = {
  problem: {
    id: 1,
    title: 'Two Sum',
    statement: 'Given an array of integers $nums$ and an integer $target$, return indices of the two numbers such that they add up to target.',
    timeLimitInSecond: 2.0,
    memoryLimitInMb: 256,
    difficulty: 'EASY',
    tags: ['ARRAY', 'STRING'],
    sampleTestCases: [
      {
        input: '[2,7,11,15]\n9',
        expectedOutput: '[0,1]',
        explanation: 'Because $nums[0] + nums[1] = 2 + 7 = 9$, we return $[0, 1]$.'
      }
    ],
    solvedStatus: 'UNATTEMPTED'
  }
};

const mockSubmissionResponse = {
  submission: {
    id: 1,
    problemSummary: {
      id: 1,
      title: 'Two Sum',
      difficulty: 'EASY',
      tags: ['ARRAY', 'STRING'],
      solvedStatus: 'UNATTEMPTED'
    },
    userSummary: {
      id: 'user1',
      handle: 'testuser',
      firstName: 'Test',
      lastName: 'User'
    },
    status: 'WAITING_FOR_EXECUTION',
    language: 'CPP',
    submittedAt: new Date().toISOString(),
    code: 'test code',
    testResults: []
  }
};

const setupHandlers = () => {
  server.use(
    http.get('http://localhost:8080/api/v1/problems/1', () => {
      return HttpResponse.json({
        problemDetails: mockProblemDetails.problem
      });
    }),
    http.post('http://localhost:8080/api/v1/submissions', async ({ request }) => {
      const body = await request.json() as { code: string; language: string; problemId: number };
      return HttpResponse.json({
        submissionDetails: {
          ...mockSubmissionResponse.submission,
          code: body.code,
          language: body.language
        }
      });
    })
  );
};

const mockedUseAuth = vi.mocked(useAuth);
const mockedUseTheme = vi.mocked(useTheme);
const mockedUseDocumentTitle = vi.mocked(useDocumentTitle);

describe('ProblemDetailPage', () => {
  beforeEach(() => {
    setupHandlers();
    mockUseParams.mockReturnValue({ id: '1' });
    mockedUseAuth.mockReturnValue({
      user: null,
      isAuthenticated: false
    });
    mockedUseTheme.mockReturnValue({
      theme: 'light'
    });
    mockedUseDocumentTitle.mockImplementation((title: string) => {
      document.title = title;
    });
  });

  describe('Loading and Error States', () => {
    it('shows loading state while fetching problem', async () => {
      server.use(
        http.get('http://localhost:8080/api/v1/problems/1', () => {
          return new Promise(() => {}); // Never resolves
        })
      );

      renderWithProviders(<ProblemDetailPage />);
      
      expect(screen.getByText('Loading problem...')).toBeInTheDocument();
      expect(screen.getByRole('status')).toBeInTheDocument();
    });

    it('shows error state when problem not found', async () => {
      server.use(
        http.get('http://localhost:8080/api/v1/problems/1', () => {
          return new HttpResponse(null, { status: 404 });
        })
      );

      renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(screen.getByText('Problem Not Found')).toBeInTheDocument();
      });
      
      expect(screen.getByText(/doesn't exist or has been removed/)).toBeInTheDocument();
    });

    it('redirects to problems list when invalid ID provided', () => {
      mockUseParams.mockReturnValue({ id: 'invalid' });

      renderWithProviders(<ProblemDetailPage />);
      
      const navigate = screen.getByTestId('navigate');
      expect(navigate).toHaveAttribute('data-to', '/problems');
      expect(navigate).toHaveAttribute('data-replace', 'true');
    });
  });

  describe('Problem Details Display', () => {
    it('displays problem information correctly', async () => {
      renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(screen.getByText(/Two Sum/)).toBeInTheDocument();
      });

      expect(screen.getByText('EASY')).toBeInTheDocument();
      expect(screen.getByText('2s')).toBeInTheDocument();
      expect(screen.getByText('256MB')).toBeInTheDocument();
      expect(screen.getByText('ARRAY')).toBeInTheDocument();
      expect(screen.getByText('STRING')).toBeInTheDocument();
    });

    it('renders LaTeX in problem statement', async () => {
      const { container } = renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(screen.getByText(/Given an array of integers/)).toBeInTheDocument();
      });

      // LaTeX rendering creates spans with katex classes
      const mathElements = container.querySelectorAll('.katex');
      expect(mathElements.length).toBeGreaterThan(0);
    });

    it('displays sample test cases with LaTeX explanations', async () => {
      const { container } = renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(screen.getByText('Example 1')).toBeInTheDocument();
      });

      expect(screen.getByText(/\[2,7,11,15\]/)).toBeInTheDocument();
      expect(screen.getByText(/\[0,1\]/)).toBeInTheDocument();
      
      // Check for LaTeX in explanation
      const explanationMath = container.querySelector('.katex');
      expect(explanationMath).toBeInTheDocument();
    });

    it('shows correct solved status icon for authenticated users', async () => {
      const solvedProblem = {
        ...mockProblemDetails,
        problem: {
          ...mockProblemDetails.problem,
          solvedStatus: 'SOLVED' as SolvedStatus
        }
      };

      server.use(
        http.get('http://localhost:8080/api/v1/problems/1', () => {
          return HttpResponse.json({
            problemDetails: solvedProblem.problem
          });
        })
      );

      mockedUseAuth.mockReturnValue({
        user: { id: 'user1', handle: 'testuser', firstName: 'Test', lastName: 'User' },
        isAuthenticated: true
      });

      renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        const solvedIcon = screen.getByTestId('CheckCircle2');
        expect(solvedIcon).toBeInTheDocument();
      });
    });

    it('hides solved status icon for unauthenticated users', async () => {
      const solvedProblem = {
        ...mockProblemDetails,
        problem: {
          ...mockProblemDetails.problem,
          solvedStatus: 'SOLVED' as SolvedStatus
        }
      };

      server.use(
        http.get('http://localhost:8080/api/v1/problems/1', () => {
          return HttpResponse.json({
            problemDetails: solvedProblem.problem
          });
        })
      );

      mockedUseAuth.mockReturnValue({
        user: null,
        isAuthenticated: false
      });

      renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(screen.getByText(/Two Sum/)).toBeInTheDocument();
      });

      const solvedIcon = screen.queryByTestId('CheckCircle2');
      expect(solvedIcon).not.toBeInTheDocument();
    });
  });

  describe('Code Editor Integration', () => {
    it('renders Monaco Editor with default C++ code', async () => {
      renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(screen.getByTestId('monaco-editor')).toBeInTheDocument();
      });

      const editor = screen.getByTestId('monaco-editor');
      expect(editor).toHaveAttribute('data-language', 'cpp');
      expect(editor.value).toContain('#include <iostream>');
    });

    it('changes language and updates boilerplate code', async () => {
      const user = userEvent.setup();
      renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(screen.getByTestId('monaco-editor')).toBeInTheDocument();
      });

      const languageSelect = screen.getByDisplayValue('C++');
      await user.click(languageSelect);
      await user.selectOptions(languageSelect, 'JAVA');

      await waitFor(() => {
        const editor = screen.getByTestId('monaco-editor');
        expect(editor).toHaveAttribute('data-language', 'java');
        expect(editor.value).toContain('public class Solution');
      });
    });

    it('updates theme based on current theme context', async () => {
      mockedUseTheme.mockReturnValue({
        theme: 'dark'
      });

      renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        const editor = screen.getByTestId('monaco-editor');
        expect(editor).toHaveAttribute('data-theme', 'vs-dark');
      });
    });

    it('allows code editing', async () => {
      const user = userEvent.setup();
      renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(screen.getByTestId('monaco-editor')).toBeInTheDocument();
      });

      const editor = screen.getByTestId('monaco-editor');
      await user.clear(editor);
      await user.type(editor, 'custom code');

      expect(editor.value).toBe('custom code');
    });
  });

  describe('Code Submission', () => {
    it('shows login prompt for unauthenticated users', async () => {
      mockedUseAuth.mockReturnValue({
        user: null,
        isAuthenticated: false
      });

      renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(screen.getByText(/Two Sum/)).toBeInTheDocument();
      });
      
      expect(screen.getByText(/login/)).toBeInTheDocument();
      expect(screen.getByText(/submit.*solution/)).toBeInTheDocument();

      expect(screen.queryByText('Submit')).not.toBeInTheDocument();
    });

    it('shows submit button for authenticated users', async () => {
      mockedUseAuth.mockReturnValue({
        user: { id: 'user1', handle: 'testuser', firstName: 'Test', lastName: 'User' },
        isAuthenticated: true
      });

      renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(screen.getByText('Submit')).toBeInTheDocument();
      });
    });

    it('submits code successfully', async () => {
      const user = userEvent.setup();
      mockedUseAuth.mockReturnValue({
        user: { id: 'user1', handle: 'testuser', firstName: 'Test', lastName: 'User' },
        isAuthenticated: true
      });

      renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(screen.getByTestId('monaco-editor')).toBeInTheDocument();
      });

      const editor = screen.getByTestId('monaco-editor');
      await user.clear(editor);
      await user.click(editor);
      await user.paste('int main() { return 0; }');

      const submitButton = screen.getByText('Submit');
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('Code submitted successfully!')).toBeInTheDocument();
      });
    });

    it('prevents submission with empty code', async () => {
      const user = userEvent.setup();
      mockedUseAuth.mockReturnValue({
        user: { id: 'user1', handle: 'testuser', firstName: 'Test', lastName: 'User' },
        isAuthenticated: true
      });

      renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(screen.getByTestId('monaco-editor')).toBeInTheDocument();
      });

      const editor = screen.getByTestId('monaco-editor');
      await user.clear(editor);

      const submitButton = screen.getByText('Submit');
      expect(submitButton).toBeDisabled();
    });

    it('handles submission errors', async () => {
      server.use(
        http.post('http://localhost:8080/api/v1/submissions', () => {
          return HttpResponse.json(
            { message: 'Rate limit exceeded' },
            { status: 429 }
          );
        })
      );

      const user = userEvent.setup();
      mockedUseAuth.mockReturnValue({
        user: { id: 'user1', handle: 'testuser', firstName: 'Test', lastName: 'User' },
        isAuthenticated: true
      });

      renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(screen.getByTestId('monaco-editor')).toBeInTheDocument();
      });

      const editor = screen.getByTestId('monaco-editor');
      await user.type(editor, 'test code');

      const submitButton = screen.getByText('Submit');
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('Rate limit exceeded')).toBeInTheDocument();
      });
    });

    it('shows loading state during submission', async () => {
      server.use(
        http.post('http://localhost:8080/api/v1/submissions', () => {
          return new Promise(() => {}); // Never resolves
        })
      );

      const user = userEvent.setup();
      mockedUseAuth.mockReturnValue({
        user: { id: 'user1', handle: 'testuser', firstName: 'Test', lastName: 'User' },
        isAuthenticated: true
      });

      renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(screen.getByTestId('monaco-editor')).toBeInTheDocument();
      });

      const editor = screen.getByTestId('monaco-editor');
      await user.type(editor, 'test code');

      const submitButton = screen.getByText('Submit');
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText('Submitting...')).toBeInTheDocument();
      });
    });
  });

  describe('Responsive Layout', () => {
    it('stacks panels vertically on mobile', async () => {
      const { container } = renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(screen.getByText(/Two Sum/)).toBeInTheDocument();
      });
      
      const grid = container.querySelector('.grid');
      expect(grid).toHaveClass('grid-cols-1', 'lg:grid-cols-2');
    });
  });

  describe('Document Title', () => {
    it('sets correct document title when problem loads', async () => {
      renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(document.title).toBe('Two Sum - Online Judge');
      });
    });

    it('sets default title when problem is loading', () => {
      server.use(
        http.get('http://localhost:8080/api/v1/problems/1', () => {
          return new Promise(() => {}); // Never resolves
        })
      );

      renderWithProviders(<ProblemDetailPage />);
      
      expect(document.title).toBe('Problem - Online Judge');
    });
  });

  describe('Accessibility', () => {
    it('has proper ARIA labels and roles', async () => {
      renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: /Two Sum/i })).toBeInTheDocument();
      });

      expect(screen.getByRole('combobox', { name: /language/i })).toBeInTheDocument();
    });

    it('supports keyboard navigation', async () => {
      const user = userEvent.setup();
      mockedUseAuth.mockReturnValue({
        user: { id: 'user1', handle: 'testuser', firstName: 'Test', lastName: 'User' },
        isAuthenticated: true
      });

      renderWithProviders(<ProblemDetailPage />);
      
      await waitFor(() => {
        expect(screen.getByTestId('monaco-editor')).toBeInTheDocument();
      });

      // Tab to language selector
      await user.tab();
      expect(screen.getByDisplayValue('C++')).toHaveFocus();

      // Tab to submit button
      await user.tab();
      expect(screen.getByText('Submit')).toHaveFocus();
    });
  });
});