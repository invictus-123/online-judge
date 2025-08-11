import { useEffect } from 'react';

export const useDocumentTitle = (title: string) => {
  useEffect(() => {
    document.title = title ? `${title} - Online Judge` : 'Online Judge';
  }, [title]);
};