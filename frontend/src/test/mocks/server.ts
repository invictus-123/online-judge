import { setupServer } from 'msw/node'
import { http, HttpResponse } from 'msw'

export const handlers = [
  http.post('/api/v1/auth/login', () => {
    return HttpResponse.json({
      token: 'mock-jwt-token',
      user: {
        id: 1,
        handle: 'testuser',
        email: 'test@example.com',
        role: 'USER'
      }
    })
  }),

  http.post('/api/v1/auth/register', () => {
    return HttpResponse.json({
      token: 'mock-jwt-token',
      user: {
        id: 1,
        handle: 'testuser',
        email: 'test@example.com',
        role: 'USER'
      }
    })
  }),

  http.post('/api/v1/auth/logout', () => {
    return new HttpResponse(null, { status: 200 })
  }),

  http.get('/api/v1/problems/list', ({ request }) => {
    const url = new URL(request.url)
    const page = Number(url.searchParams.get('page')) || 0
    const size = Number(url.searchParams.get('size')) || 10

    return HttpResponse.json({
      content: [
        {
          id: 1,
          title: 'Two Sum',
          difficulty: 'EASY',
          tags: ['ARRAY', 'HASH_TABLE'],
          solvedStatus: 'SOLVED'
        },
        {
          id: 2,
          title: 'Binary Search',
          difficulty: 'MEDIUM',
          tags: ['BINARY_SEARCH'],
          solvedStatus: 'ATTEMPTED'
        }
      ],
      totalElements: 2,
      totalPages: 1,
      size,
      number: page
    })
  }),

  http.get('/api/v1/problems/:id', ({ params }) => {
    return HttpResponse.json({
      id: Number(params.id),
      title: 'Two Sum',
      difficulty: 'EASY',
      tags: ['ARRAY', 'HASH_TABLE'],
      statement: 'Given an array of integers nums and an integer target...',
      memoryLimitMb: 256,
      timeLimitSeconds: 1,
      testCases: [
        {
          input: '[2,7,11,15]\n9',
          output: '[0,1]',
          explanation: 'nums[0] + nums[1] = 2 + 7 = 9'
        }
      ]
    })
  }),

  http.get('/api/v1/submissions/list', ({ request }) => {
    const url = new URL(request.url)
    const page = Number(url.searchParams.get('page')) || 0
    const size = Number(url.searchParams.get('size')) || 10

    return HttpResponse.json({
      content: [
        {
          id: 1,
          problemId: 1,
          problemTitle: 'Two Sum',
          difficulty: 'EASY',
          tags: ['ARRAY', 'HASH_TABLE'],
          language: 'PYTHON',
          userHandle: 'testuser',
          submittedAt: '2024-01-01T10:00:00Z',
          status: 'PASSED'
        }
      ],
      totalElements: 1,
      totalPages: 1,
      size,
      number: page
    })
  }),

  http.get('/api/v1/submissions/:id', ({ params }) => {
    return HttpResponse.json({
      id: Number(params.id),
      problemId: 1,
      problemTitle: 'Two Sum',
      difficulty: 'EASY',
      tags: ['ARRAY', 'HASH_TABLE'],
      language: 'PYTHON',
      userHandle: 'testuser',
      submittedAt: '2024-01-01T10:00:00Z',
      status: 'PASSED',
      code: 'def two_sum(nums, target):\n    return [0, 1]',
      executionTimeMs: 100,
      memoryUsageMb: 50,
      testResults: {
        totalTestCases: 2,
        passedTestCases: 2,
        results: [
          { passed: true, executionTimeMs: 50 },
          { passed: true, executionTimeMs: 50 }
        ]
      }
    })
  }),

  http.post('/api/v1/submissions', () => {
    return HttpResponse.json({
      submissionId: 1
    })
  })
]

export const server = setupServer(...handlers)