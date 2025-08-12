import { useCallback } from 'react';
import { useParams, Navigate, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Editor } from '@monaco-editor/react';
import { InlineMath, BlockMath } from 'react-katex';
import toast from 'react-hot-toast';
import { problemsService, submissionsService } from '../../services';
import { useAuth } from '../../hooks/useAuth';
import { useDocumentTitle } from '../../hooks/useDocumentTitle';
import { useTheme } from '../../hooks/useTheme';
import { useCodePersistence } from '../../hooks/useCodePersistence';
import { Button, Card, Select } from '../../components/ui';
import { SubmissionLanguage, SolvedStatus, type SubmitCodeRequest } from '../../types/api';
import { 
  Clock, 
  MemoryStick, 
  Tag, 
  Trophy,
  CheckCircle2, 
  XCircle, 
  Circle,
  Play
} from 'lucide-react';
import 'katex/dist/katex.min.css';

const LANGUAGE_OPTIONS = [
  { value: SubmissionLanguage.CPP, label: 'C++' },
  { value: SubmissionLanguage.JAVA, label: 'Java' },
  { value: SubmissionLanguage.PYTHON, label: 'Python' },
  { value: SubmissionLanguage.JAVASCRIPT, label: 'JavaScript' }
];


const getDifficultyColor = (difficulty: string) => {
  switch (difficulty) {
    case 'EASY':
      return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300';
    case 'MEDIUM':
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300';
    case 'HARD':
      return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300';
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300';
  }
};

const getSolvedStatusIcon = (status: SolvedStatus) => {
  switch (status) {
    case SolvedStatus.SOLVED:
      return <CheckCircle2 className="w-5 h-5 text-green-600" data-testid="CheckCircle2" />;
    case SolvedStatus.FAILED_ATTEMPT:
      return <XCircle className="w-5 h-5 text-red-600" data-testid="XCircle" />;
    default:
      return <Circle className="w-5 h-5 text-gray-400" data-testid="Circle" />;
  }
};

const renderLatexText = (text: string) => {
  // First, handle already properly formatted LaTeX (with $ delimiters)
  const parts = text.split(/(\$\$[\s\S]*?\$\$|\$[^$\n]*\$)/);
  
  return parts.map((part, index) => {
    // Create a unique key based on content and position
    const uniqueKey = `${part.slice(0, 20)}-${index}`;
    
    // Handle block math
    if (part.startsWith('$$') && part.endsWith('$$')) {
      const latex = part.slice(2, -2);
      return <BlockMath key={uniqueKey} math={latex} />;
    }
    // Handle inline math
    else if (part.startsWith('$') && part.endsWith('$') && part.length > 2) {
      const latex = part.slice(1, -1);
      return <InlineMath key={uniqueKey} math={latex} />;
    }
    // Handle raw LaTeX commands (convert to inline math)
    else if (part.includes('\\')) {
      // Check if this part contains LaTeX commands
      const hasLatexCommands = /\\[a-zA-Z]+/.test(part);
      if (hasLatexCommands) {
        // Wrap the entire part in inline math if it contains LaTeX
        return <InlineMath key={uniqueKey} math={part} />;
      }
      return <span key={uniqueKey}>{part}</span>;
    }
    // Regular text
    else {
      return <span key={uniqueKey}>{part}</span>;
    }
  });
};

export function ProblemDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuth();
  const { theme } = useTheme();
  const queryClient = useQueryClient();

  const problemId = id ? parseInt(id, 10) : null;
  const validProblemId = problemId && !isNaN(problemId) ? problemId : 1;

  const {
    selectedLanguage,
    code,
    setSelectedLanguage,
    updateCode
  } = useCodePersistence(validProblemId);

  const { data: problemData, isLoading, error } = useQuery({
    queryKey: ['problem', validProblemId],
    queryFn: () => problemsService.getById(validProblemId),
    enabled: !!problemId && !isNaN(problemId)
  });

  const problem = problemData?.problem;
  
  useDocumentTitle(problem ? `${problem.title} - Online Judge` : 'Problem - Online Judge');

  const submitMutation = useMutation({
    mutationFn: (request: SubmitCodeRequest) => submissionsService.submit(request),
    onSuccess: () => {
      toast.success('Code submitted successfully!');
      queryClient.invalidateQueries({ queryKey: ['problem', validProblemId] });
    },
    onError: (error: unknown) => {
      let errorMessage = 'Failed to submit code';
      if (error && typeof error === 'object' && 'response' in error) {
        const response = (error as { response?: { data?: { message?: string } } }).response;
        if (response?.data?.message) {
          errorMessage = response.data.message;
        }
      }
      toast.error(errorMessage);
    }
  });

  const handleLanguageChange = useCallback((event: React.ChangeEvent<HTMLSelectElement>) => {
    const language = event.target.value as SubmissionLanguage;
    setSelectedLanguage(language);
  }, [setSelectedLanguage]);

  const handleCodeChange = useCallback((value: string | undefined) => {
    if (value !== undefined) {
      updateCode(value);
    }
  }, [updateCode]);

  const handleSubmit = useCallback(() => {
    if (!user) {
      toast.error('Please login to submit code');
      return;
    }

    if (!code.trim()) {
      toast.error('Please write some code before submitting');
      return;
    }

    if (!problemId || isNaN(problemId)) {
      return;
    }

    submitMutation.mutate({
      problemId,
      code: code.trim(),
      language: selectedLanguage
    });
  }, [user, code, problemId, selectedLanguage, submitMutation]);

  if (!problemId || isNaN(problemId)) {
    return <Navigate to="/problems" replace />;
  }

  if (isLoading) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <output className="text-gray-600 dark:text-gray-400" aria-label="loading">Loading problem...</output>
        </div>
      </div>
    );
  }

  if (error || !problem) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">Problem Not Found</h2>
          <p className="text-gray-600 dark:text-gray-400 mb-4">
            The problem you're looking for doesn't exist or has been removed.
          </p>
          <Button onClick={() => window.history.back()}>Go Back</Button>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-6">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 min-h-[calc(100vh-200px)]">
        <div className="flex flex-col">
          <Card className="flex-1 p-6">
            <div className="mb-6">
              <div className="flex items-center gap-3 mb-4">
                {user && getSolvedStatusIcon(problem.solvedStatus)}
                <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
                  {problem.id}. {problem.title}
                </h1>
              </div>
              
              <div className="flex flex-wrap items-center gap-3 text-sm">
                <span className={`px-2 py-1 rounded-full font-medium ${getDifficultyColor(problem.difficulty)}`}>
                  <Trophy className="w-4 h-4 inline mr-1" />
                  {problem.difficulty}
                </span>
                
                <span className="flex items-center text-gray-600 dark:text-gray-400">
                  <Clock className="w-4 h-4 mr-1" />
                  {problem.timeLimitInSecond}s
                </span>
                
                <span className="flex items-center text-gray-600 dark:text-gray-400">
                  <MemoryStick className="w-4 h-4 mr-1" />
                  {problem.memoryLimitInMb}MB
                </span>
              </div>
              
              {problem.tags.length > 0 && (
                <div className="flex flex-wrap gap-2 mt-3">
                  {problem.tags.map((tag) => (
                    <span
                      key={tag}
                      className="inline-flex items-center px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300 rounded-full"
                    >
                      <Tag className="w-3 h-3 mr-1" />
                      {tag}
                    </span>
                  ))}
                </div>
              )}
            </div>

            <div className="prose prose-gray dark:prose-invert max-w-none">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                Problem Statement
              </h3>
              <div className="text-gray-700 dark:text-gray-300 whitespace-pre-wrap">
                {renderLatexText(problem.statement)}
              </div>
            </div>

            {problem.sampleTestCases.length > 0 && (
              <div className="mt-6">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                  Sample Test Cases
                </h3>
                <div className="space-y-4">
                  {problem.sampleTestCases.map((testCase, index) => (
                    <div key={`testcase-${testCase.input}-${testCase.expectedOutput}-${index}`} className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                      <h4 className="font-medium text-gray-900 dark:text-white mb-2">
                        Example {index + 1}
                      </h4>
                      
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                          <div className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">
                            Input:
                          </div>
                          <pre className="bg-white dark:bg-gray-900 p-2 rounded border text-sm font-mono overflow-x-auto">
                            {testCase.input}
                          </pre>
                        </div>
                        
                        <div>
                          <div className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">
                            Output:
                          </div>
                          <pre className="bg-white dark:bg-gray-900 p-2 rounded border text-sm font-mono overflow-x-auto">
                            {testCase.expectedOutput}
                          </pre>
                        </div>
                      </div>
                      
                      {testCase.explanation && (
                        <div className="mt-3">
                          <div className="text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">
                            Explanation:
                          </div>
                          <div className="text-sm text-gray-700 dark:text-gray-300">
                            {renderLatexText(testCase.explanation)}
                          </div>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </Card>
        </div>

        <div className="flex flex-col">
          <Card className="flex-1 p-6">
            <div className="flex flex-col h-full">
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
                  Solution
                </h2>
                <div className="flex items-center gap-3">
                  <Select
                    value={selectedLanguage}
                    onChange={handleLanguageChange}
                    options={LANGUAGE_OPTIONS}
                    className="w-32"
                    aria-label="language"
                  />
                  {user && (
                    <Button
                      onClick={handleSubmit}
                      disabled={submitMutation.isPending || !code.trim()}
                      className="flex items-center gap-2"
                    >
                      <Play className="w-4 h-4" />
                      {submitMutation.isPending ? 'Submitting...' : 'Submit'}
                    </Button>
                  )}
                </div>
              </div>

              <div className="flex-1 border rounded-lg overflow-hidden">
                <Editor
                  height="100%"
                  language={selectedLanguage.toLowerCase() === 'cpp' ? 'cpp' : selectedLanguage.toLowerCase()}
                  value={code}
                  onChange={handleCodeChange}
                  theme={theme === 'dark' ? 'vs-dark' : 'light'}
                  options={{
                    minimap: { enabled: false },
                    fontSize: 14,
                    lineNumbers: 'on',
                    roundedSelection: false,
                    scrollBeyondLastLine: false,
                    automaticLayout: true,
                    tabSize: 2,
                    insertSpaces: true,
                    wordWrap: 'on',
                    contextmenu: true,
                    selectOnLineNumbers: true
                  }}
                />
              </div>

              {!user && (
                <div className="mt-4 p-3 bg-gray-100 dark:bg-gray-800 rounded-lg border border-gray-300 dark:border-gray-600">
                  <p className="text-sm text-gray-800 dark:text-gray-200 text-center">
                    Please{' '}
                    <Link 
                      to={`/auth/login?redirect=${encodeURIComponent(window.location.pathname)}`} 
                      className="underline font-medium text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300"
                    >
                      login
                    </Link>{' '}
                    to submit your solution
                  </p>
                </div>
              )}
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
}