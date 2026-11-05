import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { PageHeader } from '@/platform/components/PageHeader';
import { renderWithProviders } from '@/test/renderWithProviders';

describe('PageHeader', () => {
  it('عنوان، زیرعنوان و اقدام اضافه را نشان می‌دهد', () => {
    renderWithProviders(
      <PageHeader title="کاربران" subtitle="فهرست کاربران" extra={<button type="button">ایجاد</button>} />,
      { route: '/users' },
    );

    expect(screen.getByRole('heading', { level: 3, name: 'کاربران' })).toBeInTheDocument();
    expect(screen.getByText('فهرست کاربران')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ایجاد' })).toBeInTheDocument();
  });
});
