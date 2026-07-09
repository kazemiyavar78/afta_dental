import { useMutation, type UseMutationOptions } from '@tanstack/react-query';
import { message } from 'antd';
import type { FieldValues, UseFormSetError, Path } from 'react-hook-form';
import type { AppError } from '../api/errorTypes';

type MutationOptions<TData, TVariables> = UseMutationOptions<TData, AppError, TVariables>;

type UseApiMutationOptions<TData, TVariables, TForm extends FieldValues> = MutationOptions<
  TData,
  TVariables
> & {
  successMessage?: string;
  setError?: UseFormSetError<TForm>;
};

/**
 * Wrapper روی useMutation با مدیریت خطا و ست کردن خطا روی فیلدهای فرم.
 */
export function useApiMutation<TData, TVariables, TForm extends FieldValues = FieldValues>(
  options: UseApiMutationOptions<TData, TVariables, TForm>,
) {
  const { successMessage, setError, onError, onSuccess, ...rest } = options;

  return useMutation<TData, AppError, TVariables>({
    ...rest,
    onSuccess: (data, variables, onMutateResult, context) => {
      if (successMessage) {
        message.success(successMessage);
      }
      onSuccess?.(data, variables, onMutateResult, context);
    },
    onError: (error, variables, onMutateResult, context) => {
      if (error.fields && setError) {
        Object.entries(error.fields).forEach(([field, msg]) => {
          setError(field as Path<TForm>, { message: msg });
        });
      } else if (!error.fields) {
        message.error(error.message);
      }
      onError?.(error, variables, onMutateResult, context);
    },
  });
}
