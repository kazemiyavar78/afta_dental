import { describe, expect, it } from 'vitest';
import { toRequestParams } from './query';

describe('toRequestParams', () => {
  it('فیلتر متن، enum، تاریخ و مرتب‌سازی را با نام پارامتر سرور می‌سازد', () => {
    expect(
      toRequestParams({
        page: 2,
        page_size: 50,
        sort: 'created_at',
        order: 'desc',
        filters: {
          actor_username: 'ali',
          event_type: 'LOGIN_FAILED',
          created_at_from: '2026-11-05',
          created_at_to: '2026-11-05',
          ip: '   ',
        },
      }),
    ).toEqual({
      page: 2,
      page_size: 50,
      sort: 'created_at',
      order: 'desc',
      'filter[actor_username]': 'ali',
      'filter[created_at_from]': '2026-11-05',
      'filter[created_at_to]': '2026-11-05',
      'filter[event_type]': 'LOGIN_FAILED',
    });
  });
});
