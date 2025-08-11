import { RegisterForm } from '../../components/auth/RegisterForm';
import { useDocumentTitle } from '../../hooks';

export const RegisterPage = () => {
  useDocumentTitle('Register');
  return <RegisterForm />;
};