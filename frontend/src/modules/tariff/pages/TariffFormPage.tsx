import { Navigate } from 'react-router-dom';

/** مسیر قدیمی ایجاد تعرفه به صفحه واحد تعرفه هدایت می‌شود. */
export function TariffFormPage() {
  return <Navigate to="/tariff" replace />;
}
