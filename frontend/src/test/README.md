# Frontend Testing Guide

This directory contains the comprehensive test suite for the DSA Platform frontend application.

## Test Structure

```
src/test/
├── README.md                 # This file
├── setup.ts                 # Test environment setup
├── utils/
│   └── test-utils.tsx       # Custom render function with providers
├── mocks/
│   └── server.ts            # MSW server for API mocking
└── index.test.ts           # Health check test
```

## Testing Stack

- **Vitest**: Fast unit test runner built on Vite
- **@testing-library/react**: Simple and complete testing utilities
- **@testing-library/user-event**: Fire events for user interaction testing
- **MSW (Mock Service Worker)**: API mocking for integration tests
- **jsdom**: Browser environment simulation for Node.js

## Available Scripts

```bash
# Run tests in watch mode
npm run test

# Run tests with UI dashboard
npm run test:ui

# Run tests with coverage report
npm run test:coverage
```

## Test Categories

### 1. Component Tests
- **Authentication Components**: `LoginForm`, `RegisterForm`, `ProtectedRoute`
- **UI Components**: `Button`, `Card`, `Input`, `Select`
- **Layout Components**: `Navbar`, `Layout`
- **Page Components**: `HomePage`, problem and submission pages

### 2. Service Tests
- **API Services**: `auth-service`, `problems-service`, `submissions-service`
- **Utility Functions**: Token storage, API client configuration

### 3. Hook Tests
- **Custom Hooks**: `useAuth`, `useTheme`
- **Context Integration**: Provider integration tests

## Testing Patterns

### Component Testing with Providers
```tsx
import { render } from '../test/utils/test-utils'
import { MyComponent } from './MyComponent'

test('renders component correctly', () => {
  render(<MyComponent />)
  // Test assertions...
})
```

### API Mocking
```tsx
import { server } from '../test/mocks/server'
import { http, HttpResponse } from 'msw'

test('handles API error', async () => {
  server.use(
    http.post('/api/v1/auth/login', () => {
      return new HttpResponse(null, { status: 401 })
    })
  )
  
  // Test component behavior with mocked API error...
})
```

### User Interaction Testing
```tsx
import userEvent from '@testing-library/user-event'

test('handles user input', async () => {
  const user = userEvent.setup()
  render(<LoginForm />)
  
  await user.type(screen.getByLabelText(/handle/i), 'testuser')
  await user.click(screen.getByRole('button', { name: /sign in/i }))
  
  // Assert expected behavior...
})
```

## Mock Services

The test suite includes comprehensive API mocking for:
- Authentication endpoints (`/api/v1/auth/*`)
- Problems endpoints (`/api/v1/problems/*`)
- Submissions endpoints (`/api/v1/submissions/*`)

All mock responses match the actual API contract and include realistic data.

## Coverage Goals

The test suite aims for:
- **90%+ statement coverage** for critical paths
- **Complete component coverage** for all UI components
- **Service layer coverage** for all API interactions
- **Hook coverage** for all custom hooks

## Running Tests in CI/CD

Tests are automatically run in GitHub Actions when:
- Changes are made to the `frontend/` directory
- Pull requests are created or updated
- Pushes are made to the main branch

Coverage reports are generated and deployed to GitHub Pages.

## Best Practices

1. **Test Behavior, Not Implementation**: Focus on what the user sees and does
2. **Use Semantic Queries**: Prefer `getByRole`, `getByLabelText` over `getByTestId`
3. **Mock External Dependencies**: Use MSW for API calls, mock complex libraries
4. **Test Error States**: Include tests for loading, error, and edge cases
5. **Keep Tests Isolated**: Each test should be independent and repeatable
6. **Use Real User Interactions**: Use `user-event` instead of `fireEvent`

## Debugging Tests

```bash
# Run specific test file
npm test -- LoginForm.test.tsx

# Run tests in debug mode
npm test -- --reporter=verbose

# Open Vitest UI for interactive debugging
npm run test:ui
```

## Adding New Tests

1. Create test files adjacent to components: `Component.test.tsx`
2. Import from test utilities: `import { render } from '../test/utils/test-utils'`
3. Use descriptive test names that explain the behavior being tested
4. Group related tests using `describe` blocks
5. Clean up after tests using `beforeEach` and `afterEach` as needed