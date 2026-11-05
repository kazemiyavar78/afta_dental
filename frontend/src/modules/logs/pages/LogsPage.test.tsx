import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '@/test/renderWithProviders';
import { fetchLogIntegrity, fetchLogMeta, fetchLogs } from '../api';
import { LogsPage } from './LogsPage';

vi.mock('../api', () => ({
  fetchLogs: vi.fn(),
  fetchLogMeta: vi.fn(),
  fetchLogIntegrity: vi.fn(),
}));

const logRow = {
  id: 9,
  created_at: '2026-11-05T08:30:00Z',
  actor_username: 'sara',
  target_name: 'ali',
  target_type: 'user',
  event_type: 'USER_UPDATED',
  event_type_label: 'به‌روزرسانی کاربر',
  result: 'success',
  reason: 'role_changed',
  reason_label: 'تغییر نقش',
  ip: '192.168.1.4',
  user_agent: 'Mozilla/5.0 Chrome/120.0.0.0',
  description: 'به‌روزرسانی کاربر',
  details: { name: { before: 'قدیم', after: 'جدید' } },
};

describe('LogsPage', () => {
  beforeEach(() => {
    vi.mocked(fetchLogs).mockResolvedValue({ items: [logRow], total: 1, page: 1, page_size: 20 });
    vi.mocked(fetchLogMeta).mockResolvedValue({
      event_types: [{ value: 'USER_UPDATED', label: 'به‌روزرسانی کاربر' }],
      reasons: [{ value: 'role_changed', label: 'تغییر نقش' }],
      target_types: [{ value: 'user', label: 'کاربر' }],
    });
    vi.mocked(fetchLogIntegrity).mockResolvedValue({
      last_check_at: '2026-11-05T09:00:00Z',
      last_result: 'ok',
      broken_at_id: 0,
      open_findings_count: 0,
      open_findings: [],
    });
  });

  it('ستون شناسه کاربر ندارد و وضعیت سالم را با زمان بررسی نشان می‌دهد', async () => {
    renderWithProviders(<LogsPage />, { route: '/logs' });

    expect(await screen.findByText('سالم')).toBeInTheDocument();
    expect(screen.getByText(/آخرین بررسی:/)).toBeInTheDocument();
    const headers = screen.getAllByRole('columnheader').map((cell) => cell.textContent ?? '');
    expect(headers.some((text) => text.includes('شناسه کاربر'))).toBe(false);
    expect(headers.some((text) => text.includes('نام کاربری انجام‌دهنده'))).toBe(true);
    expect(screen.getByText('sara')).toBeInTheDocument();
    expect(screen.getByText('موفق')).toBeInTheDocument();
    expect(screen.getByText('Chrome')).toBeInTheDocument();
    expect(screen.queryByText('4242')).not.toBeInTheDocument();
  });

  it('وضعیت خراب را با شناسه اولین ردیف خراب نشان می‌دهد', async () => {
    vi.mocked(fetchLogIntegrity).mockResolvedValue({
      last_check_at: '2026-11-05T09:00:00Z',
      last_result: 'broken',
      broken_at_id: 42,
      open_findings_count: 0,
      open_findings: [],
    });
    renderWithProviders(<LogsPage />, { route: '/logs' });
    expect(await screen.findByText('خراب')).toBeInTheDocument();
    expect(screen.getByText('اولین ردیف خراب: 42')).toBeInTheDocument();
    await waitFor(() => expect(fetchLogs).toHaveBeenCalled());
  });

  it('تعداد یافته‌های باز را نشان می‌دهد و با کلیک فهرست را باز می‌کند', async () => {
    vi.mocked(fetchLogIntegrity).mockResolvedValue({
      last_check_at: '2026-11-05T09:00:00Z',
      last_result: 'ok',
      broken_at_id: 0,
      open_findings_count: 2,
      open_findings: [
        {
          entity_type: 'user',
          entity_type_label: 'کاربر',
          entity_name: 'mina-broken',
          first_seen_at: '2026-11-05T08:00:00Z',
        },
        {
          entity_type: 'role',
          entity_type_label: 'نقش',
          entity_name: 'role-broken',
          first_seen_at: '2026-11-05T08:05:00Z',
        },
      ],
    });
    renderWithProviders(<LogsPage />, { route: '/logs' });
    fireEvent.click(await screen.findByRole('button', { name: 'یافته‌های باز: 2' }));
    expect(await screen.findByText(/mina-broken/)).toBeInTheDocument();
    expect(screen.getByText(/role-broken/)).toBeInTheDocument();
    expect(screen.getByText('یافته‌های باز یکپارچگی')).toBeInTheDocument();
  });
});
