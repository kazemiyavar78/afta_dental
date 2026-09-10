# Report Inventory

> آخرین بررسی: ۱۴۰۵/۰۶/۱۶ — Phase 1 Audit

## خلاصه

| معیار | مقدار |
|---|---|
| **تعداد Report/Print قابل migration** | 4 |
| **صفحات Report اختصاصی** | 2 (با 3 زیرنوع در Doctor Performance) |
| **PDF** | ✅ `@react-pdf/renderer` |
| **Excel** | ✅ `xlsx` |
| **Print** | ✅ HTML engine مرکزی |

---

## جدول Inventory

| Report | Location | Data Source | PDF | Excel | Print | Page Size | Orientation | Status |
|---|---|---|---|---|---|---|---|---|
| گزارش درآمد به تفکیک سازمان | `modules/reports/pages/RevenueByOrganizationPage.tsx` | `GET /reports/revenue-by-organization` | ✅ | ✅ | ✅ | A4 | portrait | **Migrated** |
| گزارش کارکرد — تفکیک سازمان | `modules/reports/pages/DoctorPerformancePage.tsx` | `GET /reports/doctor-performance/by-organization` | ✅ | ✅ | ✅ | A4 | portrait | **Migrated** |
| گزارش کارکرد — تفکیک خدمت | همان صفحه | `GET /reports/doctor-performance/by-service` | ✅ | ✅ | ✅ | A4 | portrait | **Migrated** |
| گزارش کارکرد — لیست بیماران | همان صفحه | `GET /reports/doctor-performance/patient-list` | ✅ | ✅ | ✅ | A4 | portrait | **Migrated** |
| فهرست بیماران | `modules/patients/pages/PatientsPage.tsx` | `GET /patients` | ✅ | ✅ | ✅ | A4 | portrait | **Migrated** |
| قبض پذیرش | `modules/reception/printReceptionReceipt.ts` | Zustand store | ❌ | ❌ | ✅ | 70mm | portrait | **Partial** |

---

## موارد غیر-Report (Table بدون Export/Print)

| Component | Location | Print | Export | یادداشت |
|---|---|---|---|---|
| DataTable عمومی | `platform/components/DataTable` | ❌ | ❌ | Pagination کلاینت‌سایd (10/20/50) |
| دفترچه کیف پول بیمار | `wallet/components/PatientWalletLedgerModal.tsx` | ❌ | ❌ | Modal با فیلتر |
| تاریخچه خدمات بیمار | `reception/components/PatientServicesHistoryModal.tsx` | ❌ | ❌ | Modal |
| پذیرش‌های بیمار | `reception/components/PatientReceptionsModal.tsx` | ❌ | ❌ | Modal |
| سایر صفحات CRUD | users, organization, services, … | ❌ | ❌ | لیست‌های مدیریتی |

---

## الگوی Print فعلی (مشترک)

همه implementationهای print از الگوی زیر پیروی می‌کنند:

1. ساخت HTML string با template literal
2. CSS inline شامل `@page`, `@media print`, RTL
3. `window.open('', '_blank')` → `document.write(html)` → `window.print()`
4. فونت: `Tahoma, "Segoe UI", sans-serif`
5. تاریخ شمسی: `dayjs().calendar('jalali').locale('fa')`
6. اعداد فارسی: `toLocaleString('fa-IR')`

### CSS Print مشترک

| فایل | `@page size` | margin | thead repeat | summary row |
|---|---|---|---|---|
| `printRevenueByOrganizationA4.ts` | A4 portrait | 12mm | ✅ | ✅ |
| `printDoctorPerformanceA4.ts` | A4 portrait | 12mm | ✅ | ✅ + sections |
| `printPatientsA4.ts` | A4 portrait | 12mm | ✅ | ❌ |
| `printReceptionReceipt.ts` | 70mm auto | 2mm | ❌ | N/A (receipt) |

---

## APIهای Report (Backend)

| Endpoint | Permission | Pagination |
|---|---|---|
| `GET /reports/revenue-by-organization` | `reports.read` | ❌ — تمام داده یکجا |
| `GET /reports/doctor-performance/by-organization` | `reports.read` | ❌ |
| `GET /reports/doctor-performance/by-service` | `reports.read` | ❌ |
| `GET /reports/doctor-performance/patient-list` | `reports.read` | ❌ |

> **نکته Export:** Backend pagination ندارد؛ Export = همان dataset فیلترشده API.

---

## Formatterهای تکراری

| Function | محل‌ها |
|---|---|
| `formatMoney` | `reports/types.ts`, `printReceptionReceipt.ts`, wallet/reception components |
| `escapeHtml` | هر 4 فایل print |
| `formatGregorianDate` | 3 فایل print A4 |
| `formatBirthDate` | `printPatientsA4.ts` |

---

## Dependencyهای فعلی

```json
// frontend/package.json — مرتبط با Report
{
  "dayjs": "^1.11.21",
  "jalaliday": "^3.1.1",
  "antd": "^6.5.0"
}
```

**نیاز به نصب:**
- `@react-pdf/renderer` — PDF
- `xlsx` — Excel
- `vitest` + `@testing-library/react` + `jsdom` — تست (فعلاً test runner وجود ندارد)

---

## ریسک‌های Migration

| ریسک | شدت | راهکار |
|---|---|---|
| Doctor Performance — 3 نوع column + sections + grand total | بالا | `ReportSection` + dynamic definition |
| قبض پذیرش — layout غیرجدولی 70mm | متوسط | `DocumentPrintTemplate` جدا از tabular engine |
| فونت فارسی PDF | متوسط | Register Vazirmatn در `@react-pdf/renderer` |
| Patients — UI pagination vs print all | پایین | API بدون pagination؛ print از `data` کامل |
| Duplicate CSS/HTML در 4 فایل | پایین | Centralize در `platform/reports/exporters/print.ts` |

---

## تصمیم معماری (Phase 2)

- **مسیر Engine:** `frontend/src/platform/reports/` (نه `shared/` — مطابق convention پروژه)
- **Pilot Migration:** RevenueByOrganization (ساده‌ترین — یک جدول + summary)
- **Reception Receipt:** migration جزئی — reuse formatters + print utilities، layout custom بماند
