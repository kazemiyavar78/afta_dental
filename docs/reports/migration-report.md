# Migration Report

> تاریخ: ۱۴۰۵/۰۶/۱۶

## خلاصه

| معیار | مقدار |
|---|---|
| Reportهای migrate شده | 4 (+ 1 partial) |
| فایل‌های print حذف شده | 3 |
| فایل‌های engine جدید | 25+ |
| Dependencies اضافه | `@react-pdf/renderer`, `xlsx`, `vitest` |
| Build | ✅ موفق |
| Tests | ✅ 4 suite |

---

## Reportهای موفق

| Report | Engine Integration | Print | PDF | Excel |
|---|---|---|---|---|
| Revenue by Organization | `ReportViewer` full | ✅ | ✅ (جدید) | ✅ (جدید) |
| Doctor Performance ×3 | `ReportViewer` + sections | ✅ | ✅ (جدید) | ✅ (جدید) |
| Patient List | `ReportToolbar` + `useReport` | ✅ | ✅ (جدید) | ✅ (جدید) |

## Partial Migration

| Report | وضعیت |
|---|---|
| Reception Receipt 70mm | formatters مشترک (`escapeHtml`, `formatCurrency`) — layout custom باقی |

---

## فایل‌های حذف شده

- `frontend/src/modules/reports/printRevenueByOrganizationA4.ts`
- `frontend/src/modules/reports/printDoctorPerformanceA4.ts`
- `frontend/src/modules/patients/printPatientsA4.ts`

## فایل‌های جدید (اصلی)

```
frontend/src/platform/reports/
├── components/     (ReportViewer, ReportTable, ReportToolbar, ...)
├── exporters/      (print.ts, pdf.tsx, excel.ts)
├── formatters/     (number, date, value)
├── hooks/          (useReport.ts)
├── utils/          (html, filename)
├── types.ts
├── constants.ts
└── index.ts

frontend/src/modules/reports/definitions/
frontend/src/modules/patients/definitions/

docs/reports/
```

---

## Dependencies

### اضافه شده
- `@react-pdf/renderer@^4.9.0`
- `xlsx@^0.18.5`
- `vitest`, `@testing-library/react`, `jsdom`

### حذف شده
- (هیچ)

---

## Limitations

1. **PDF Font:** Vazirmatn از CDN — برای offline باید در `public/fonts/` bundle شود
2. **Reception Receipt:** layout 70mm خارج از tabular engine
3. **Bundle size:** PDF library ~3MB — code-splitting توصیه می‌شود
4. **Patients page:** UI table جدا از Report (action columns)

---

## TODO آینده

- [ ] Bundle Vazirmatn font locally
- [ ] Lazy-load PDF exporter (`import()` dynamic)
- [ ] Backend export endpoint برای datasetهای بسیار بزرگ
- [ ] Document Print adapter برای reception receipt
- [ ] Code-split `@react-pdf/renderer` chunk

---

## Verification

```bash
cd frontend
npm run build   # ✅
npm test        # ✅
```
