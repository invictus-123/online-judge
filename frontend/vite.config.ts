import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'

export default defineConfig(() => {
  const gcsBucketUrl = process.env.VITE_GCS_BUCKET_URL || './';

  return {
    plugins: [react()],
    base: gcsBucketUrl,
  };
});
