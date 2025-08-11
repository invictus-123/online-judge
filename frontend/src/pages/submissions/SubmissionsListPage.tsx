import { useDocumentTitle } from '../../hooks';

export const SubmissionsListPage = () => {
  useDocumentTitle('Submissions');
  
  return (
    <div className="flex min-h-[50vh] items-center justify-center">
      <div className="text-center">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Submissions Page</h2>
        <p className="text-gray-600 dark:text-gray-400">Coming soon...</p>
      </div>
    </div>
  );
};