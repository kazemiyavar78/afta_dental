export type WalletBalance = {
  file_id: number;
  balance: number;
};

export type WalletTransaction = {
  id: number;
  file_id: number;
  reception_id?: number | null;
  amount: number;
  category: string;
  action: string;
  payment_method?: string | null;
  description: string;
  bank_account_id?: number | null;
  counterparty_card?: string | null;
  tracking_number?: string | null;
  bank_name?: string | null;
  credit_account_id?: number | null;
  credit_account_name?: string | null;
  /** فقط برای کسر از اعتبار: آیا مبلغ واقعاً پرداخت شده است */
  credit_paid?: boolean;
  credit_paid_at?: string | null;
  paid_at?: string | null;
  /** شناسه کاربر عامل؛ 0 یعنی سیستم */
  performed_by: number;
  /** نام نمایشی عامل (سیستم یا نام کاربر) */
  performed_by_name: string;
  created_at: string;
};

export type CashAdjustPayload = {
  file_id: number;
  amount: number;
  increase: boolean;
  description?: string;
};

export type CardToCardAdjustPayload = {
  file_id: number;
  amount: number;
  increase: boolean;
  bank_account_id: number;
  counterparty_card: string;
  tracking_number: string;
  paid_at?: string | null;
  description?: string;
};

export type CardReaderChargePayload = {
  file_id: number;
  amount: number;
  transaction_number: string;
  card_number: string;
  approved: boolean;
  description?: string;
};

export type BankAccount = {
  id: number;
  bank_name: string;
  sheba_number: string;
  account_number: string;
  card_number: string;
  account_name: string;
  has_transactions: boolean;
};

export type BankAccountPayload = {
  bank_name: string;
  sheba_number: string;
  account_number: string;
  card_number: string;
  account_name: string;
};

export type CreditChargePayload = {
  file_id: number;
  credit_account_id: number;
  amount: number;
  description?: string;
};

export type CreditAccount = {
  id: number;
  name: string;
  description: string;
  credit_limit: number;
  used_amount: number;
  remaining: number;
  is_active: boolean;
  has_transactions: boolean;
};

export type CreditAccountPayload = {
  name: string;
  description: string;
  credit_limit: number;
  is_active: boolean;
};

export type CardReaderSimulateResponse = {
  transaction_number: string;
  card_number: string;
  amount: number;
  approved: boolean;
  paid_at: string;
};
