import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'

export default defineConfig(() => {
  const gcsBucketUrl = process.env.VITE_GCS_BUCKET_URL || './';
  
  console.log('Vite config - VITE_GCS_BUCKET_URL:', process.env.VITE_GCS_BUCKET_URL);
  console.log('Vite config - Using base URL:', gcsBucketUrl);

  return {
    plugins: [react()],
    base: gcsBucketUrl,
  };
});
