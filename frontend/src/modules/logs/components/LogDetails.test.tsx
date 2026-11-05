import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { LogDetails } from './LogDetails';

describe('LogDetails', () => {
  it('اختلاف فیلد و مجوزها را نشان می‌دهد و JSON خام و شناسه کاربر را نه', () => {
    render(
      <LogDetails
        details={{
          name: { before: 'قدیم', after: 'جدید' },
          user_id: 4242,
          added_permissions: ['roles.read'],
          removed_permissions: ['logs.read'],
        }}
      />,
    );

    expect(screen.getByText('فیلد')).toBeInTheDocument();
    expect(screen.getByText('قبل')).toBeInTheDocument();
    expect(screen.getByText('بعد')).toBeInTheDocument();
    expect(screen.getByText('نام')).toBeInTheDocument();
    expect(screen.getByText('قدیم')).toBeInTheDocument();
    expect(screen.getByText('جدید')).toBeInTheDocument();
    expect(screen.getByText('اضافه‌شده')).toBeInTheDocument();
    expect(screen.getByText('roles.read')).toBeInTheDocument();
    expect(screen.getByText('حذف‌شده')).toBeInTheDocument();
    expect(screen.getByText('logs.read')).toBeInTheDocument();
    expect(screen.queryByText('4242')).not.toBeInTheDocument();
    expect(screen.queryByText(/\{/)).not.toBeInTheDocument();
  });

  it('تغییر تنظیم را یک ردیف قبل و بعد نشان می‌دهد', () => {
    render(<LogDetails details={{ key: 'Password_MinLength', before: '8', after: '10' }} />);
    expect(screen.getByText('Password_MinLength')).toBeInTheDocument();
    expect(screen.getByText('8')).toBeInTheDocument();
    expect(screen.getByText('10')).toBeInTheDocument();
  });
});
