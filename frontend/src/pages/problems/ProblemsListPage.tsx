import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { Card, CardContent, CardTitle } from '../../components/ui/Card';
import { Button } from '../../components/ui/Button';
import { Select } from '../../components/ui/Select';
import { problemsService } from '../../services/problems-service';
import { useAuth } from '../../hooks/useAuth';
import type { 
  ProblemDifficulty, 
  ProblemTag, 
  SolvedStatus,
  ProblemFilterRequest 
} from '../../types/api';

const DIFFICULTY_OPTIONS = [
  { value: 'EASY', label: 'Easy' },
  { value: 'MEDIUM', label: 'Medium' },
  { value: 'HARD', label: 'Hard' }
] as const;

const TAG_OPTIONS = [
  { value: 'ARRAY', label: 'Array' },
  { value: 'STRING', label: 'String' },
  { value: 'GREEDY', label: 'Greedy' },
  { value: 'DP', label: 'Dynamic Programming' },
  { value: 'TREE', label: 'Tree' },
  { value: 'GRAPH', label: 'Graph' }
] as const;

const SOLVED_STATUS_OPTIONS = [
  { value: 'SOLVED', label: 'Solved' },
  { value: 'ATTEMPTED', label: 'Attempted' },
  { value: 'UNATTEMPTED', label: 'Unattempted' }
] as const;

export function ProblemsListPage() {
  const { isAuthenticated } = useAuth();
  const [currentPage, setCurrentPage] = useState(1);
  const [selectedDifficulties, setSelectedDifficulties] = useState<ProblemDifficulty[]>([]);
  const [selectedTags, setSelectedTags] = useState<ProblemTag[]>([]);
  const [selectedSolvedStatuses, setSelectedSolvedStatuses] = useState<SolvedStatus[]>([]);

  const filters = useMemo<ProblemFilterRequest>(() => ({
    page: currentPage,
    difficulties: selectedDifficulties.length > 0 ? selectedDifficulties : undefined,
    tags: selectedTags.length > 0 ? selectedTags : undefined,
    solvedStatuses: selectedSolvedStatuses.length > 0 ? selectedSolvedStatuses : undefined,
  }), [currentPage, selectedDifficulties, selectedTags, selectedSolvedStatuses]);

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['problems', filters],
    queryFn: () => problemsService.list(filters),
  });

  const handleDifficultyChange = (value: string) => {
    if (value && value !== selectedDifficulties[0]) {
      setSelectedDifficulties([value as ProblemDifficulty]);
      setCurrentPage(1);
    } else if (!value) {
      setSelectedDifficulties([]);
      setCurrentPage(1);
    }
  };

  const handleTagChange = (value: string) => {
    if (value && !selectedTags.includes(value as ProblemTag)) {
      setSelectedTags([...selectedTags, value as ProblemTag]);
      setCurrentPage(1);
    }
  };

  const handleSolvedStatusChange = (value: string) => {
    if (value && value !== selectedSolvedStatuses[0]) {
      setSelectedSolvedStatuses([value as SolvedStatus]);
      setCurrentPage(1);
    } else if (!value) {
      setSelectedSolvedStatuses([]);
      setCurrentPage(1);
    }
  };

  const removeTag = (tagToRemove: ProblemTag) => {
    setSelectedTags(selectedTags.filter(tag => tag !== tagToRemove));
    setCurrentPage(1);
  };

  const clearAllFilters = () => {
    setSelectedDifficulties([]);
    setSelectedTags([]);
    setSelectedSolvedStatuses([]);
    setCurrentPage(1);
  };

  const getSolvedStatusIcon = (status: SolvedStatus) => {
    switch (status) {
      case 'SOLVED':
        return <span className="text-green-500 text-lg">✅</span>;
      case 'ATTEMPTED':
        return <span className="text-red-500 text-lg">❌</span>;
      default:
        return null;
    }
  };

  const getDifficultyColor = (difficulty: ProblemDifficulty) => {
    switch (difficulty) {
      case 'EASY':
        return 'text-green-600 bg-green-50 border-green-200 dark:text-green-400 dark:bg-green-900/20 dark:border-green-800';
      case 'MEDIUM':
        return 'text-yellow-600 bg-yellow-50 border-yellow-200 dark:text-yellow-400 dark:bg-yellow-900/20 dark:border-yellow-800';
      case 'HARD':
        return 'text-red-600 bg-red-50 border-red-200 dark:text-red-400 dark:bg-red-900/20 dark:border-red-800';
      default:
        return 'text-gray-600 bg-gray-50 border-gray-200 dark:text-gray-400 dark:bg-gray-900/20 dark:border-gray-800';
    }
  };

  const formatTagDisplay = (tag: ProblemTag) => {
    return TAG_OPTIONS.find(option => option.value === tag)?.label || tag;
  };

  const renderProblemsContent = () => {
    if (isLoading) {
      const skeletonItems = Array.from({ length: 6 }, (_, i) => `skeleton-${Date.now()}-${i}`);
      
      return (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {skeletonItems.map((skeletonId) => (
            <Card key={skeletonId} className="animate-pulse">
              <CardContent className="p-6">
                <div className="h-6 bg-gray-200 dark:bg-gray-700 rounded mb-2"></div>
                <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded mb-4"></div>
                <div className="flex gap-2 mb-2">
                  <div className="h-6 w-12 bg-gray-200 dark:bg-gray-700 rounded"></div>
                  <div className="h-6 w-16 bg-gray-200 dark:bg-gray-700 rounded"></div>
                </div>
                <div className="flex gap-1">
                  <div className="h-5 w-8 bg-gray-200 dark:bg-gray-700 rounded"></div>
                  <div className="h-5 w-12 bg-gray-200 dark:bg-gray-700 rounded"></div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      );
    }

    if (data && data.problems.length > 0) {
      return (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {data.problems.map(problem => (
            <Link key={problem.id} to={`/problems/${problem.id}`}>
              <Card className="h-full transition-colors hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer">
                <CardContent className="p-6">
                  <div className="flex items-start justify-between mb-2">
                    <CardTitle className="text-lg leading-tight">
                      {problem.id}. {problem.title}
                    </CardTitle>
                    {isAuthenticated && getSolvedStatusIcon(problem.solvedStatus)}
                  </div>
                  
                  <div className="flex items-center gap-2 mb-3">
                    <span className={`px-2 py-1 rounded-md text-xs font-medium border ${getDifficultyColor(problem.difficulty)}`}>
                      {problem.difficulty}
                    </span>
                  </div>

                  <div className="flex flex-wrap gap-1">
                    {problem.tags.slice(0, 3).map(tag => (
                      <span
                        key={tag}
                        className="px-2 py-1 text-xs rounded-md bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300"
                      >
                        {formatTagDisplay(tag)}
                      </span>
                    ))}
                    {problem.tags.length > 3 && (
                      <span className="px-2 py-1 text-xs rounded-md bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300">
                        +{problem.tags.length - 3}
                      </span>
                    )}
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      );
    }

    return (
      <div className="flex flex-col items-center justify-center min-h-[30vh]">
        <div className="text-center">
          <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">
            No problems found
          </h3>
          <p className="text-gray-600 dark:text-gray-400 mb-4">
            Try adjusting your filters or check back later for new problems.
          </p>
          {(selectedDifficulties.length > 0 || selectedTags.length > 0 || selectedSolvedStatuses.length > 0) && (
            <Button variant="outline" onClick={clearAllFilters}>
              Clear All Filters
            </Button>
          )}
        </div>
      </div>
    );
  };

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="flex flex-col items-center justify-center min-h-[50vh]">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
              Error Loading Problems
            </h2>
            <p className="text-gray-600 dark:text-gray-400 mb-6">
              Something went wrong while fetching the problems.
            </p>
            <Button onClick={() => refetch()}>
              Try Again
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
          Problems
        </h1>
        <p className="text-gray-600 dark:text-gray-400">
          Solve coding challenges and improve your skills
        </p>
      </div>

      <div className="mb-6 space-y-4">
        <div className="flex flex-wrap gap-4">
          <div className="min-w-[150px]">
            <Select
              value={selectedDifficulties[0] || ''}
              onChange={(e) => handleDifficultyChange(e.target.value)}
              placeholder="Difficulty"
              options={[
                { value: '', label: 'All Difficulties' },
                ...DIFFICULTY_OPTIONS
              ]}
            />
          </div>

          <div className="min-w-[150px]">
            <Select
              value=""
              onChange={(e) => handleTagChange(e.target.value)}
              placeholder="Add Tag"
              options={[
                { value: '', label: 'Select Tag' },
                ...TAG_OPTIONS.filter(option => !selectedTags.includes(option.value as ProblemTag))
              ]}
            />
          </div>

          {isAuthenticated && (
            <div className="min-w-[150px]">
              <Select
                value={selectedSolvedStatuses[0] || ''}
                onChange={(e) => handleSolvedStatusChange(e.target.value)}
                placeholder="Status"
                options={[
                  { value: '', label: 'All Status' },
                  ...SOLVED_STATUS_OPTIONS
                ]}
              />
            </div>
          )}

          {(selectedDifficulties.length > 0 || selectedTags.length > 0 || selectedSolvedStatuses.length > 0) && (
            <Button variant="outline" size="sm" onClick={clearAllFilters}>
              Clear Filters
            </Button>
          )}
        </div>

        {selectedTags.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {selectedTags.map(tag => (
              <span
                key={tag}
                className="inline-flex items-center gap-1 px-3 py-1 rounded-full text-sm bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
              >
                {formatTagDisplay(tag)}
                <button
                  onClick={() => removeTag(tag)}
                  className="ml-1 text-blue-600 hover:text-blue-800 dark:text-blue-300 dark:hover:text-blue-100"
                >
                  ×
                </button>
              </span>
            ))}
          </div>
        )}
      </div>

      {renderProblemsContent()}

      {data && data.problems.length > 0 && (
        <div className="mt-8 flex justify-center gap-2">
          <Button
            variant="outline"
            disabled={currentPage === 1}
            onClick={() => setCurrentPage(currentPage - 1)}
          >
            Previous
          </Button>
          <span className="flex items-center px-4 py-2 text-sm text-gray-700 dark:text-gray-300">
            Page {currentPage}
          </span>
          <Button
            variant="outline"
            disabled={!data || data.problems.length < 20}
            onClick={() => setCurrentPage(currentPage + 1)}
          >
            Next
          </Button>
        </div>
      )}
    </div>
  );
}