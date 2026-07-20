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

export type CardReaderSimulateResponse = {
  transaction_number: string;
  card_number: string;
  amount: number;
  approved: boolean;
  paid_at: string;
};
