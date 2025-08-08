import { Link } from 'react-router-dom';
import { Code2, FileText, BarChart3, Users, Trophy } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle, Button } from '../components/ui';
import { useAuth } from '../hooks';

export const HomePage = () => {
  const { isAuthenticated, user } = useAuth();

  return (
    <div className="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <div className="text-center">
        <div className="mb-8 inline-flex items-center justify-center rounded-full bg-blue-100 p-3 dark:bg-blue-900/20">
          <Code2 className="h-12 w-12 text-blue-600 dark:text-blue-400" />
        </div>
        <h1 className="mb-6 text-4xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-6xl">
          Online Judge
        </h1>
        <p className="mx-auto mb-8 max-w-2xl text-xl text-gray-600 dark:text-gray-300">
          Master Data Structures & Algorithms with our comprehensive online judge platform. 
          Practice and improve your coding skills.
        </p>

        {isAuthenticated ? (
          <div className="mb-12">
            <h2 className="mb-4 text-2xl font-semibold text-gray-900 dark:text-white">
              Welcome back, {user?.handle}! 👋
            </h2>
            <p className="text-gray-600 dark:text-gray-300">
              Ready to solve some problems?
            </p>
          </div>
        ) : (
          <div className="mb-12 space-x-4">
            <Link to="/auth/register">
              <Button size="lg" className="px-8">
                Get Started
              </Button>
            </Link>
            <Link to="/auth/login">
              <Button variant="outline" size="lg" className="px-8">
                Sign In
              </Button>
            </Link>
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-3">
        <Card className="text-center">
          <CardHeader>
            <div className="mb-4 inline-flex items-center justify-center rounded-full bg-green-100 p-3 dark:bg-green-900/20">
              <FileText className="h-8 w-8 text-green-600 dark:text-green-400" />
            </div>
            <CardTitle>Practice Problems</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="mb-4 text-gray-600 dark:text-gray-300">
              Solve hundreds of carefully curated problems ranging from easy to hard difficulty levels.
            </p>
            <Link to="/problems">
              <Button variant="outline" className="w-full">
                Browse Problems
              </Button>
            </Link>
          </CardContent>
        </Card>

        <Card className="text-center">
          <CardHeader>
            <div className="mb-4 inline-flex items-center justify-center rounded-full bg-purple-100 p-3 dark:bg-purple-900/20">
              <BarChart3 className="h-8 w-8 text-purple-600 dark:text-purple-400" />
            </div>
            <CardTitle>Track Progress</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="mb-4 text-gray-600 dark:text-gray-300">
              Monitor your submissions, track your progress, and see detailed results for each attempt.
            </p>
            <Link to="/submissions">
              <Button variant="outline" className="w-full">
                View Submissions
              </Button>
            </Link>
          </CardContent>
        </Card>

        <Card className="text-center">
          <CardHeader>
            <div className="mb-4 inline-flex items-center justify-center rounded-full bg-orange-100 p-3 dark:bg-orange-900/20">
              <Code2 className="h-8 w-8 text-orange-600 dark:text-orange-400" />
            </div>
            <CardTitle>Multiple Languages</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="mb-4 text-gray-600 dark:text-gray-300">
              Code in C++, Java, Python, or JavaScript. Practice in your preferred programming language.
            </p>
            {isAuthenticated ? (
              <Link to="/problems">
                <Button variant="outline" className="w-full">
                  Start Coding
                </Button>
              </Link>
            ) : (
              <Link to="/auth/register">
                <Button variant="outline" className="w-full">
                  Join Now
                </Button>
              </Link>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="mt-16">
        <h2 className="mb-8 text-center text-3xl font-bold text-gray-900 dark:text-white">
          Platform Statistics
        </h2>
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
          <div className="text-center">
            <div className="mb-2 inline-flex items-center justify-center rounded-full bg-blue-100 p-3 dark:bg-blue-900/20">
              <FileText className="h-6 w-6 text-blue-600 dark:text-blue-400" />
            </div>
            <div className="text-2xl font-bold text-gray-900 dark:text-white">500+</div>
            <div className="text-sm text-gray-600 dark:text-gray-400">Problems Available</div>
          </div>
          <div className="text-center">
            <div className="mb-2 inline-flex items-center justify-center rounded-full bg-green-100 p-3 dark:bg-green-900/20">
              <Users className="h-6 w-6 text-green-600 dark:text-green-400" />
            </div>
            <div className="text-2xl font-bold text-gray-900 dark:text-white">10K+</div>
            <div className="text-sm text-gray-600 dark:text-gray-400">Active Users</div>
          </div>
          <div className="text-center">
            <div className="mb-2 inline-flex items-center justify-center rounded-full bg-purple-100 p-3 dark:bg-purple-900/20">
              <BarChart3 className="h-6 w-6 text-purple-600 dark:text-purple-400" />
            </div>
            <div className="text-2xl font-bold text-gray-900 dark:text-white">1M+</div>
            <div className="text-sm text-gray-600 dark:text-gray-400">Submissions</div>
          </div>
          <div className="text-center">
            <div className="mb-2 inline-flex items-center justify-center rounded-full bg-orange-100 p-3 dark:bg-orange-900/20">
              <Trophy className="h-6 w-6 text-orange-600 dark:text-orange-400" />
            </div>
            <div className="text-2xl font-bold text-gray-900 dark:text-white">4</div>
            <div className="text-sm text-gray-600 dark:text-gray-400">Languages Supported</div>
          </div>
        </div>
      </div>

      {!isAuthenticated && (
        <div className="mt-16 rounded-lg bg-blue-600 px-8 py-12 text-center dark:bg-blue-700">
          <h2 className="mb-4 text-3xl font-bold text-white">
            Ready to Start Your Journey?
          </h2>
          <p className="mb-8 text-xl text-blue-100">
            Join thousands of developers improving their coding skills every day.
          </p>
          <div className="space-x-4">
            <Link to="/auth/register">
              <Button variant="outline" size="lg" className="bg-white text-blue-600 hover:bg-gray-100">
                Sign Up Now
              </Button>
            </Link>
            <Link to="/problems">
              <Button variant="ghost" size="lg" className="text-white hover:bg-blue-500">
                Browse Problems
              </Button>
            </Link>
          </div>
        </div>
      )}
    </div>
  );
};