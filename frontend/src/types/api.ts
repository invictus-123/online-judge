export const ProblemDifficulty = {
  EASY: 'EASY',
  MEDIUM: 'MEDIUM',
  HARD: 'HARD'
} as const;

export type ProblemDifficulty = typeof ProblemDifficulty[keyof typeof ProblemDifficulty];

export const ProblemTag = {
  ARRAY: 'ARRAY',
  STRING: 'STRING',
  GREEDY: 'GREEDY',
  DP: 'DP',
  TREE: 'TREE',
  GRAPH: 'GRAPH'
} as const;

export type ProblemTag = typeof ProblemTag[keyof typeof ProblemTag];

export const SolvedStatus = {
  SOLVED: 'SOLVED',
  FAILED_ATTEMPT: 'FAILED_ATTEMPT',
  UNATTEMPTED: 'UNATTEMPTED'
} as const;

export type SolvedStatus = typeof SolvedStatus[keyof typeof SolvedStatus];

export const SubmissionLanguage = {
  CPP: 'CPP',
  JAVA: 'JAVA',
  PYTHON: 'PYTHON',
  JAVASCRIPT: 'JAVASCRIPT'
} as const;

export type SubmissionLanguage = typeof SubmissionLanguage[keyof typeof SubmissionLanguage];

export const SubmissionStatus = {
  WAITING_FOR_EXECUTION: 'WAITING_FOR_EXECUTION',
  RUNNING: 'RUNNING',
  PASSED: 'PASSED',
  TIME_LIMIT_EXCEEDED: 'TIME_LIMIT_EXCEEDED',
  MEMORY_LIMIT_EXCEEDED: 'MEMORY_LIMIT_EXCEEDED',
  COMPILATION_ERROR: 'COMPILATION_ERROR',
  RUNTIME_ERROR: 'RUNTIME_ERROR'
} as const;

export type SubmissionStatus = typeof SubmissionStatus[keyof typeof SubmissionStatus];

export const UserRole = {
  USER: 'USER',
  ADMIN: 'ADMIN'
} as const;

export type UserRole = typeof UserRole[keyof typeof UserRole];

// Request DTOs
export interface LoginRequest {
  handle: string;
  password: string;
}

export interface RegisterRequest {
  handle: string;
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  userRole: UserRole;
}

export interface SubmitCodeRequest {
  problemId: number;
  code: string;
  language: SubmissionLanguage;
}

export interface ProblemFilterRequest {
  difficulties?: ProblemDifficulty[];
  tags?: ProblemTag[];
  solvedStatuses?: SolvedStatus[];
  page: number;
}

export interface SubmissionFilterRequest {
  onlyMe?: boolean;
  problemId?: number;
  statuses?: SubmissionStatus[];
  languages?: SubmissionLanguage[];
  page: number;
}

// Response DTOs
export interface AuthResponse {
  token: string;
}

export interface ErrorResponse {
  message: string;
  details?: Record<string, string>;
  timestamp: string;
  status: number;
}

// UI DTOs
export interface UserSummaryUi {
  id: string;
  handle: string;
  firstName: string;
  lastName: string;
}

export interface TestCaseUi {
  input: string;
  expectedOutput: string;
  explanation: string;
}

export interface ProblemSummaryUi {
  id: number;
  title: string;
  difficulty: ProblemDifficulty;
  tags: ProblemTag[];
  solvedStatus: SolvedStatus;
}

export interface ProblemDetailsUi {
  id: number;
  title: string;
  statement: string;
  timeLimitInSecond: number;
  memoryLimitInMb: number;
  difficulty: ProblemDifficulty;
  tags: ProblemTag[];
  sampleTestCases: TestCaseUi[];
  solvedStatus: SolvedStatus;
}

export interface TestResultSummaryUi {
  testCaseIndex: number;
  passed: boolean;
  executionTime?: number;
  memoryUsed?: number;
  actualOutput?: string;
}

export interface SubmissionSummaryUi {
  id: number;
  problemSummary: ProblemSummaryUi;
  userSummary: UserSummaryUi;
  status: SubmissionStatus;
  language: SubmissionLanguage;
  submittedAt: string;
}

export interface SubmissionDetailsUi {
  id: number;
  problemSummary: ProblemSummaryUi;
  userSummary: UserSummaryUi;
  status: SubmissionStatus;
  language: SubmissionLanguage;
  submittedAt: string;
  code: string;
  executionTime?: number;
  memoryUsed?: number;
  testResults: TestResultSummaryUi[];
}

// API Response wrappers
export interface ListProblemsResponse {
  problems: ProblemSummaryUi[];
}

export interface GetProblemByIdResponse {
  problem: ProblemDetailsUi;
}

export interface ListSubmissionsResponse {
  submissions: SubmissionSummaryUi[];
}

export interface GetSubmissionByIdResponse {
  submission: SubmissionDetailsUi;
}

export interface SubmitCodeResponse {
  submission: SubmissionSummaryUi;
}

export interface CreateProblemResponse {
  problem: ProblemSummaryUi;
}