import { httpClient } from '@/platform/api/httpClient';
import type {
  BankAccount,
  BankAccountPayload,
  CardReaderChargePayload,
  CardReaderSimulateResponse,
  CardToCardAdjustPayload,
  CashAdjustPayload,
  CreditAccount,
  CreditAccountPayload,
  CreditChargePayload,
  WalletBalance,
  WalletTransaction,
} from './types';

/** دریافت موجودی کیف پول پرونده */
export async function fetchWalletBalance(fileId: number): Promise<WalletBalance> {
  const { data } = await httpClient.get<WalletBalance>(`/wallet/${fileId}/balance`);
  return data;
}

/** دریافت تراکنش‌های پرونده */
export async function fetchWalletTransactions(fileId: number): Promise<WalletTransaction[]> {
  const { data } = await httpClient.get<WalletTransaction[]>(`/wallet/${fileId}/transactions`);
  return data;
}

/** افزایش/کاهش اعتبار نقدی */
export async function adjustCash(payload: CashAdjustPayload): Promise<WalletTransaction> {
  const { data } = await httpClient.post<WalletTransaction>('/wallet/cash', payload);
  return data;
}

/** افزایش/کاهش کارت‌به‌کارت */
export async function adjustCardToCard(payload: CardToCardAdjustPayload): Promise<WalletTransaction> {
  const { data } = await httpClient.post<WalletTransaction>('/wallet/card-to-card', payload);
  return data;
}

/** شارژ از پاسخ کارتخوان */
export async function chargeFromCardReader(payload: CardReaderChargePayload): Promise<WalletTransaction> {
  const { data } = await httpClient.post<WalletTransaction>('/wallet/card-reader', payload);
  return data;
}

/** شبیه‌سازی پاسخ کارتخوان */
export async function simulateCardReader(amount: number): Promise<CardReaderSimulateResponse> {
  const { data } = await httpClient.post<CardReaderSimulateResponse>('/card-reader/simulate', { amount });
  return data;
}

/** لیست حساب‌های بانکی */
export async function fetchBankAccounts(): Promise<BankAccount[]> {
  const { data } = await httpClient.get<BankAccount[]>('/bank-accounts');
  return data;
}

/** ایجاد حساب بانکی */
export async function createBankAccount(payload: BankAccountPayload): Promise<BankAccount> {
  const { data } = await httpClient.post<BankAccount>('/bank-accounts', payload);
  return data;
}

/** ویرایش حساب بانکی */
export async function updateBankAccount(id: number, payload: BankAccountPayload): Promise<BankAccount> {
  const { data } = await httpClient.put<BankAccount>(`/bank-accounts/${id}`, payload);
  return data;
}

/** حذف حساب بانکی */
export async function deleteBankAccount(id: number): Promise<void> {
  await httpClient.delete(`/bank-accounts/${id}`);
}

/** لیست حساب‌های اعتباری */
export async function fetchCreditAccounts(): Promise<CreditAccount[]> {
  const { data } = await httpClient.get<CreditAccount[]>('/credit-accounts');
  return data;
}

/** لیست حساب‌های اعتباری فعال برای فرم پرداخت */
export async function fetchActiveCreditAccounts(): Promise<CreditAccount[]> {
  const { data } = await httpClient.get<CreditAccount[]>('/credit-accounts/active');
  return data;
}

/** ایجاد حساب اعتباری */
export async function createCreditAccount(payload: CreditAccountPayload): Promise<CreditAccount> {
  const { data } = await httpClient.post<CreditAccount>('/credit-accounts', payload);
  return data;
}

/** ویرایش حساب اعتباری */
export async function updateCreditAccount(id: number, payload: CreditAccountPayload): Promise<CreditAccount> {
  const { data } = await httpClient.put<CreditAccount>(`/credit-accounts/${id}`, payload);
  return data;
}

/** حذف حساب اعتباری */
export async function deleteCreditAccount(id: number): Promise<void> {
  await httpClient.delete(`/credit-accounts/${id}`);
}

/** شارژ پرونده مالی از محل کسر حساب اعتباری */
export async function chargeFromCreditAccount(payload: CreditChargePayload): Promise<WalletTransaction> {
  const { data } = await httpClient.post<WalletTransaction>('/wallet/credit-charge', payload);
  return data;
}

/**
 * وضعیت پرداخت‌شدن تراکنش کسر از اعتبار را ثبت یا برمی‌دارد.
 * @param id شناسه تراکنش
 * @param creditPaid true یعنی مبلغ پرداخت شده است
 */
export async function setCreditChargePaid(id: number, creditPaid: boolean): Promise<WalletTransaction> {
  const { data } = await httpClient.patch<WalletTransaction>(`/wallet/transactions/${id}/credit-paid`, {
    credit_paid: creditPaid,
  });
  return data;
}
