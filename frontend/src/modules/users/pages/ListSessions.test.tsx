import { screen, waitFor } from '@testing-library/react';
import { Route, Routes } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '@/test/renderWithProviders';
import { fetchSessionPage, fetchUser } from '../api';
import { ListSessions } from './ListSessions';

vi.mock('../api', async () => {
  const actual = await vi.importActual<typeof import('../api')>('../api');
  return {
    ...actual,
    fetchUser: vi.fn(),
    fetchSessionPage: vi.fn(),
  };
});

describe('ListSessions', () => {
  it('شناسه عددی کاربر را نشان نمی‌دهد و فهرست را به همان کاربر محدود می‌کند', async () => {
    vi.mocked(fetchUser).mockResolvedValue({
      id: 7,
      username: 'ali',
      address: '',
      name: 'علی',
      family: 'رضایی',
      phone_number: '',
      medical_code: null,
      role_id: 1,
      user_type: 'User',
      role_name: 'پذیرش',
      is_active: true,
      is_locked: false,
      last_login_at: null,
    });
    vi.mocked(fetchSessionPage).mockResolvedValue({
      items: [
        {
          id: 'session-1',
          username: 'ali',
          ip: '10.0.0.8',
          browser: 'Chrome',
          creation_time: '2026-11-05T08:00:00Z',
          user_id: 987654,
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    });

    renderWithProviders(
      <Routes>
        <Route path="/users/:id/sessions" element={<ListSessions />} />
      </Routes>,
      { route: '/users/7/sessions' },
    );

    expect(await screen.findByText('10.0.0.8')).toBeInTheDocument();
    expect(screen.getByRole('columnheader', { name: 'نام کاربری' })).toBeInTheDocument();
    expect(screen.getByRole('columnheader', { name: 'آدرس IP' })).toBeInTheDocument();
    expect(screen.getByRole('columnheader', { name: 'مرورگر' })).toBeInTheDocument();
    expect(screen.getByRole('columnheader', { name: 'زمان ورود' })).toBeInTheDocument();
    expect(screen.getByRole('columnheader', { name: 'عملیات' })).toBeInTheDocument();
    expect(screen.queryByText('987654')).not.toBeInTheDocument();
    await waitFor(() => {
      expect(fetchSessionPage).toHaveBeenCalledWith(
        expect.objectContaining({
          filters: expect.objectContaining({ user_id_from: '7', user_id_to: '7' }),
        }),
      );
    });
  });
});
