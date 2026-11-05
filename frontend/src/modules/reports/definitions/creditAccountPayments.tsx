import { Checkbox } from 'antd';
import type { ReportDefinition, ReportMetaItem } from '@/platform/reports';
import { formatCurrency, formatDateTime, formatGregorianDate } from '@/platform/reports';
import type { CreditAccountPaymentRow, CreditAccountPaymentsMeta } from '../types';

type CreditPaidColumnOptions = {
  canMarkPaid: boolean;
  pendingId: number | null;
  onTogglePaid: (row: CreditAccountPaymentRow, paid: boolean) => void;
};

/** متن وضعیت پرداخت را برای چاپ و خروجی برمی‌گرداند. */
function creditPaidLabel(paid: boolean): string {
  return paid ? 'پرداخت شده' : 'پرداخت نشده';
}

/** meta ردیف‌های هدر گزارش پرداخت‌های اعتباری را می‌سازد. */
export function buildCreditAccountPaymentsMetaRows(meta: CreditAccountPaymentsMeta): ReportMetaItem[][] {
  return [
    [
      { label: 'از تاریخ', value: `${formatGregorianDate(meta.from_date)} — ${meta.from_time}` },
      { label: 'تا تاریخ', value: `${formatGregorianDate(meta.to_date)} — ${meta.to_time}` },
    ],
    [{ label: 'نام حساب', value: meta.credit_account_name || 'همه حساب‌ها' }],
  ];
}

/**
 * تعریف ستون‌ها و خروجی گزارش پرداخت‌های حساب اعتباری.
 * تیک پرداخت فقط در جدول صفحه است؛ چاپ و اکسل متن وضعیت را می‌نویسند.
 */
export function buildCreditAccountPaymentsDefinition(
  options: CreditPaidColumnOptions,
): ReportDefinition<CreditAccountPaymentRow> {
  return {
    id: 'credit-account-payments',
    title: 'گزارش پرداخت‌های حساب‌های اعتباری',
    description: 'پرداخت‌هایی که از محل کسر حساب اعتباری در پرونده مالی شارژ شده‌اند',
    page: { size: 'A4', orientation: 'landscape', margin: 12 },
    columns: [
      {
        key: 'paid_at',
        title: 'تاریخ',
        width: 14,
        format: 'datetime',
        formatter: (value) => formatDateTime(value as string),
      },
      { key: 'credit_account_name', title: 'نام حساب', width: 16, align: 'right', format: 'text' },
      { key: 'file_number', title: 'شماره پرونده', width: 12, align: 'right', format: 'text' },
      { key: 'patient_name', title: 'بیمار', width: 16, align: 'right', format: 'text' },
      {
        key: 'amount',
        title: 'مبلغ',
        width: 12,
        format: 'currency',
        formatter: (value) => formatCurrency(value),
      },
      { key: 'description', title: 'توضیحات', width: 20, align: 'right', format: 'text' },
      { key: 'performed_by_name', title: 'ثبت‌کننده', width: 14, align: 'right', format: 'text' },
      {
        key: 'credit_paid',
        title: 'پرداخت شده',
        width: 14,
        align: 'center',
        format: 'text',
        formatter: (_value, row) => (
          <Checkbox
            checked={row.credit_paid}
            disabled={!options.canMarkPaid || options.pendingId === row.id}
            onChange={(e) => options.onTogglePaid(row, e.target.checked)}
          >
            {row.credit_paid ? 'بله' : 'خیر'}
          </Checkbox>
        ),
        printFormatter: (value) => creditPaidLabel(Boolean(value)),
        pdfFormatter: (value) => creditPaidLabel(Boolean(value)),
        excelFormatter: (value) => creditPaidLabel(Boolean(value)),
      },
    ],
    summary: {
      label: 'جمع',
      fields: [{ key: 'amount', label: 'جمع مبلغ', format: 'currency' }],
    },
    header: { showGeneratedAt: true },
    footer: { text: 'پرداخت‌های کسر از حساب اعتباری' },
    toolbar: { print: true, pdf: true, excel: true, pageSetup: true },
    export: { sheetName: 'پرداخت اعتباری' },
  };
}
