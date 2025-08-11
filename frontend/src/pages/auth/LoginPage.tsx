import { LoginForm } from '../../components/auth/LoginForm';
import { useDocumentTitle } from '../../hooks';

export const LoginPage = () => {
  useDocumentTitle('Login');
  return <LoginForm />;
};