import dayjs from 'dayjs';
import type { Patient } from './types';

/** برچسب جنسیت برای نمایش */
function sexLabel(sex: boolean): string {
  return sex ? 'مرد' : 'زن';
}

/** تاریخ را برای چاپ به شمسی تبدیل می‌کند. */
function formatBirthDate(value: string): string {
  if (!value) return '—';
  return dayjs(value).calendar('jalali').locale('fa').format('YYYY/MM/DD');
}

/**
 * قالب چاپ A4 چندصفحه‌ای لیست فیلترشده بیماران را در پنجره جدید باز می‌کند.
 * @param patients لیست بیماران قابل چاپ
 * @param title عنوان گزارش
 */
export function printPatientsA4(patients: Patient[], title = 'فهرست بیماران'): void {
  const rows = patients
    .map(
      (p, index) => `
      <tr>
        <td>${index + 1}</td>
        <td>${p.file_number}</td>
        <td>${p.first_name}</td>
        <td>${p.last_name}</td>
        <td>${p.national_code}</td>
        <td>${formatBirthDate(p.birth_date)}</td>
        <td>${sexLabel(p.sex)}</td>
        <td>${p.mobile_phone_number ?? '—'}</td>
        <td>${p.home_phone_number ?? '—'}</td>
        <td>${p.address ?? '—'}</td>
      </tr>`,
    )
    .join('');

  const html = `<!DOCTYPE html>
<html lang="fa" dir="rtl">
<head>
  <meta charset="utf-8" />
  <title>${title}</title>
  <style>
    @page {
      size: A4 portrait;
      margin: 12mm;
    }
    * { box-sizing: border-box; }
    body {
      font-family: Tahoma, "Segoe UI", sans-serif;
      font-size: 11px;
      color: #111;
      margin: 0;
      direction: rtl;
    }
    .header {
      text-align: center;
      margin-bottom: 12px;
      border-bottom: 2px solid #222;
      padding-bottom: 8px;
    }
    .header h1 {
      margin: 0 0 4px;
      font-size: 16px;
    }
    .meta {
      display: flex;
      justify-content: space-between;
      margin-bottom: 8px;
      font-size: 10px;
      color: #444;
    }
    table {
      width: 100%;
      border-collapse: collapse;
      table-layout: fixed;
    }
    th, td {
      border: 1px solid #333;
      padding: 4px 5px;
      text-align: center;
      word-wrap: break-word;
      vertical-align: middle;
    }
    th {
      background: #efefef;
      font-weight: 700;
    }
    tr {
      page-break-inside: avoid;
    }
    thead {
      display: table-header-group;
    }
    tfoot {
      display: table-footer-group;
    }
    .footer {
      margin-top: 10px;
      font-size: 10px;
      text-align: left;
      color: #555;
    }
    @media print {
      body { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
    }
  </style>
</head>
<body>
  <div class="header">
    <h1>${title}</h1>
  </div>
  <div class="meta">
    <span>تعداد رکورد: ${patients.length}</span>
    <span>تاریخ چاپ: ${dayjs().calendar('jalali').locale('fa').format('YYYY/MM/DD HH:mm')}</span>
  </div>
  <table>
    <thead>
      <tr>
        <th style="width:4%">ردیف</th>
        <th style="width:9%">شماره پرونده</th>
        <th style="width:9%">نام</th>
        <th style="width:11%">نام خانوادگی</th>
        <th style="width:10%">کد ملی</th>
        <th style="width:9%">تاریخ تولد</th>
        <th style="width:6%">جنسیت</th>
        <th style="width:10%">موبایل</th>
        <th style="width:10%">تلفن منزل</th>
        <th style="width:22%">آدرس</th>
      </tr>
    </thead>
    <tbody>
      ${rows || '<tr><td colspan="10">موردی برای چاپ یافت نشد</td></tr>'}
    </tbody>
  </table>
  <div class="footer">صفحات به‌صورت خودکار بر اساس اندازه A4 شکسته می‌شوند.</div>
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
