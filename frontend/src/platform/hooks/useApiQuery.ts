import {
  useQuery,
  type UseQueryOptions,
  type UseQueryResult,
} from '@tanstack/react-query';
import type { AppError } from '../api/errorTypes';

type ApiQueryOptions<T> = Omit<UseQueryOptions<T, AppError>, 'queryKey' | 'queryFn'> & {
  queryKey: unknown[];
  queryFn: () => Promise<T>;
};

/**
 * Wrapper روی TanStack Query با مدیریت خطای یکسان.
 */
export function useApiQuery<T>(options: ApiQueryOptions<T>): UseQueryResult<T, AppError> {
  return useQuery<T, AppError>({
    ...options,
    retry: (failureCount, error) => {
      if (error.code === 'UNAUTHORIZED' || error.code === 'FORBIDDEN') return false;
      return failureCount < 2;
    },
  });
}
