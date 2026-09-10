import dayjs from 'dayjs';
import { escapeHtml, formatCurrency } from '@/platform/reports';
import { lineCashAmount, type PatientFormState, type ServiceLineState } from './types';

/** داده لازم برای چاپ قبض ۷ سانتی پذیرش */
export type ReceptionReceiptPrintData = {
  receptionId: number | null;
  receptionDate: string;
  patient: PatientFormState;
  insuranceName: string;
  additionalInsuranceName: string;
  specialCodeValue: string;
  specialCodeName: string;
  referralCode: number | null;
  doctorName: string;
  doctorMedicalCode: string | null;
  assistantName: string;
  description: string;
  discount: number;
  services: ServiceLineState[];
};

/** تاریخ را برای چاپ به شمسی تبدیل می‌کند. */
function formatDate(value: string | null | undefined): string {
  if (!value) return '—';
  const d = dayjs(value);
  if (!d.isValid()) return '—';
  return d.calendar('jalali').locale('fa').format('YYYY/MM/DD');
}

/** مبلغ را با جداکننده هزارگان فارسی نمایش می‌دهد. */
function formatMoney(value: number): string {
  return formatCurrency(value);
}

/** مقدار خالی را با خط تیره جایگزین می‌کند. */
function dash(value: string | number | null | undefined): string {
  const t = value == null ? '' : String(value).trim();
  return t ? escapeHtml(t) : '—';
}

/**
 * قالب چاپ قبض پذیرش با عرض ۷ سانتی‌متر را در پنجره جدید باز می‌کند.
 * شامل اطلاعات بیمار، بیمه، پزشک/دستیار، خدمات و جمع‌های مالی است.
 * @param data داده پذیرش جاری برای چاپ
 */
export function printReceptionReceipt(data: ReceptionReceiptPrintData): void {
  const patientName = [data.patient.first_name, data.patient.last_name].filter(Boolean).join(' ');
  const lines = data.services.filter((s) => s.service_id > 0);

  const totals = lines.reduce(
    (acc, line) => {
      acc.amount += Number(line.service_amount ?? 0);
      acc.tariff += Number(line.service_tariff ?? 0);
      acc.orgShare += Number(line.service_organization_share ?? 0);
      acc.suppShare += Number(line.service_supplementary_insurance_share ?? 0);
      acc.subsidy += Number(line.service_subsidy_share ?? 0);
      acc.cashBeforeDiscount += lineCashAmount(line);
      return acc;
    },
    { amount: 0, tariff: 0, orgShare: 0, suppShare: 0, subsidy: 0, cashBeforeDiscount: 0 },
  );
  const cashTotal = Math.max(0, totals.cashBeforeDiscount - Number(data.discount ?? 0));

  const specialCodeLabel =
    data.specialCodeValue || data.specialCodeName
      ? [data.specialCodeValue, data.specialCodeName].filter(Boolean).join(' — ')
      : '';

  const serviceBlocks = lines
    .map((line, index) => {
      const toothBits: string[] = [];
      if (line.teeth_number != null) toothBits.push(`دندان ${line.teeth_number}`);
      if (line.teeth_direction != null) toothBits.push(`جهت ${line.teeth_direction}`);
      const meta = [
        `×${formatMoney(line.quantity)}`,
        ...toothBits,
        line.service_description?.trim() || '',
      ]
        .filter(Boolean)
        .join(' · ');
      const cash = lineCashAmount(line);
      return `
      <div class="svc">
        <div class="svc-title">${index + 1}. ${dash(line.service_name)}${meta ? ` <span class="muted">(${escapeHtml(meta)})</span>` : ''}</div>
        <div class="grid">
          <span>نرخ</span><b>${formatMoney(line.service_amount)}</b>
          <span>تعرفه</span><b>${formatMoney(line.service_tariff)}</b>
          <span>سازمان</span><b>${formatMoney(line.service_organization_share)}</b>
          <span>تکمیلی</span><b>${formatMoney(line.service_supplementary_insurance_share)}</b>
          <span>یارانه</span><b>${formatMoney(line.service_subsidy_share)}</b>
          <span>صندوق</span><b>${formatMoney(cash)}</b>
        </div>
      </div>`;
    })
    .join('');

  const html = `<!DOCTYPE html>
<html lang="fa" dir="rtl">
<head>
  <meta charset="utf-8" />
  <title>قبض پذیرش${data.receptionId != null ? ` #${data.receptionId}` : ''}</title>
  <style>
    @page {
      size: 70mm auto;
      margin: 2mm;
    }
    * { box-sizing: border-box; }
    html, body {
      margin: 0;
      padding: 0;
      direction: rtl;
      color: #000;
      background: #fff;
      font-family: Tahoma, "Segoe UI", sans-serif;
      font-size: 9px;
      line-height: 1.35;
      -webkit-print-color-adjust: exact;
      print-color-adjust: exact;
    }
    .receipt {
      width: 66mm;
      max-width: 66mm;
      margin: 0 auto;
      padding: 1mm 0;
    }
    .title {
      text-align: center;
      border-bottom: 1.5px solid #000;
      padding-bottom: 3px;
      margin-bottom: 4px;
    }
    .title h1 {
      margin: 0;
      font-size: 12px;
      font-weight: 700;
      letter-spacing: 0.2px;
    }
    .title .sub {
      margin-top: 1px;
      font-size: 8px;
    }
    .row {
      display: flex;
      justify-content: space-between;
      gap: 4px;
      margin: 1px 0;
    }
    .row span.label { color: #222; white-space: nowrap; }
    .row b { font-weight: 700; text-align: left; word-break: break-word; }
    .section {
      border-top: 1px dashed #000;
      padding-top: 3px;
      margin-top: 4px;
    }
    .section-title {
      font-size: 9px;
      font-weight: 700;
      margin-bottom: 2px;
      text-align: center;
      border-bottom: 1px solid #000;
      padding-bottom: 1px;
    }
    .muted { color: #333; font-weight: 400; font-size: 8px; }
    .svc {
      margin: 3px 0 4px;
      padding-bottom: 3px;
      border-bottom: 1px dotted #666;
    }
    .svc:last-child { border-bottom: none; padding-bottom: 0; }
    .svc-title {
      font-weight: 700;
      margin-bottom: 2px;
      word-break: break-word;
    }
    .grid {
      display: grid;
      grid-template-columns: auto 1fr auto 1fr;
      column-gap: 3px;
      row-gap: 1px;
      font-size: 8px;
    }
    .grid span { color: #222; }
    .grid b { text-align: left; font-variant-numeric: tabular-nums; }
    .totals .grid {
      grid-template-columns: auto 1fr;
      font-size: 9px;
    }
    .pay {
      margin-top: 5px;
      border: 1.5px solid #000;
      padding: 4px 3px;
      text-align: center;
    }
    .pay .lbl {
      font-size: 9px;
      font-weight: 700;
      margin-bottom: 2px;
    }
    .pay .amt {
      font-size: 14px;
      font-weight: 800;
      letter-spacing: 0.3px;
    }
    .foot {
      margin-top: 5px;
      text-align: center;
      font-size: 7.5px;
      border-top: 1px solid #000;
      padding-top: 3px;
    }
    @media print {
      html, body { width: 70mm; }
      .receipt { width: 66mm; }
    }
  </style>
</head>
<body>
  <div class="receipt">
    <div class="title">
      <h1>قبض پذیرش</h1>
      <div class="sub">
        ${data.receptionId != null ? `شماره ${escapeHtml(data.receptionId)} · ` : ''}
        تاریخ ${escapeHtml(formatDate(data.receptionDate))}
      </div>
    </div>

    <div class="section">
      <div class="section-title">اطلاعات بیمار</div>
      <div class="row"><span class="label">نام</span><b>${dash(patientName)}</b></div>
      <div class="row"><span class="label">کد ملی</span><b>${dash(data.patient.national_code)}</b></div>
      <div class="row"><span class="label">پرونده</span><b>${dash(data.patient.file_number)}</b></div>
      <div class="row"><span class="label">موبایل</span><b>${dash(data.patient.mobile_phone_number)}</b></div>
    </div>

    <div class="section">
      <div class="section-title">بیمه و کد خاص</div>
      <div class="row"><span class="label">بیمه پایه</span><b>${dash(data.insuranceName)}</b></div>
      <div class="row"><span class="label">بیمه تکمیلی</span><b>${dash(data.additionalInsuranceName)}</b></div>
      <div class="row"><span class="label">کد خاص</span><b>${dash(specialCodeLabel)}</b></div>
      ${
        data.referralCode != null
          ? `<div class="row"><span class="label">کد ارجاع</span><b>${dash(data.referralCode)}</b></div>`
          : ''
      }
    </div>

    <div class="section">
      <div class="section-title">پزشک و دستیار</div>
      <div class="row"><span class="label">پزشک</span><b>${dash(data.doctorName)}${
        data.doctorMedicalCode
          ? ` <span class="muted">(${escapeHtml(data.doctorMedicalCode)})</span>`
          : ''
      }</b></div>
      <div class="row"><span class="label">دستیار</span><b>${dash(data.assistantName)}</b></div>
      ${
        data.description?.trim()
          ? `<div class="row"><span class="label">توضیحات</span><b>${dash(data.description)}</b></div>`
          : ''
      }
    </div>

    <div class="section">
      <div class="section-title">خدمات دریافت‌شده</div>
      ${serviceBlocks || '<div class="muted" style="text-align:center">خدمتی ثبت نشده است</div>'}
    </div>

    <div class="section totals">
      <div class="section-title">جمع‌ها</div>
      <div class="grid">
        <span>نرخ</span><b>${formatMoney(totals.amount)}</b>
        <span>تعرفه</span><b>${formatMoney(totals.tariff)}</b>
        <span>سازمان</span><b>${formatMoney(totals.orgShare)}</b>
        <span>تکمیلی</span><b>${formatMoney(totals.suppShare)}</b>
        <span>یارانه</span><b>${formatMoney(totals.subsidy)}</b>
        <span>تخفیف</span><b>${formatMoney(data.discount)}</b>
      </div>
      <div class="pay">
        <div class="lbl">مبلغ پرداختی بیمار (صندوق)</div>
        <div class="amt">${formatMoney(cashTotal)} ریال</div>
      </div>
    </div>

    <div class="foot">
      تاریخ چاپ: ${escapeHtml(dayjs().calendar('jalali').locale('fa').format('YYYY/MM/DD HH:mm'))}
    </div>
  </div>
  <script>
    window.onload = function () {
      window.focus();
      window.print();
    };
  </script>
</body>
</html>`;

  const win = window.open('', '_blank');
  if (!win) {
    throw new Error('اجازه باز شدن پنجره چاپ داده نشد');
  }
  win.document.open();
  win.document.write(html);
  win.document.close();
}
