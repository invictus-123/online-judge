import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react-swc'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const gcsBucketUrl = env.VITE_GCS_BUCKET_URL || './';

  return {
    plugins: [react()],
    base: gcsBucketUrl,
  };
});
