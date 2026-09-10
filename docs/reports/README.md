# Report Engine

Report Engine مرکزی پروژه در `frontend/src/platform/reports/` قرار دارد.

## اگر Report جدید می‌سازی

1. API را در module بساز (`modules/<name>/api.ts`)
2. types را بساز (`modules/<name>/types.ts`)
3. query hook را بساز (با `useApiQuery`)
4. `ReportDefinition` را بساز (`modules/<name>/definitions/`)
5. `ReportViewer` را در صفحه استفاده کن
6. **هیچ PDF/Excel/Print logic اختصاصی داخل Report page نساز**

## مستندات

- [architecture.md](./architecture.md)
- [report-definition.md](./report-definition.md)
- [pdf.md](./pdf.md)
- [excel.md](./excel.md)
- [print.md](./print.md)
- [migration.md](./migration.md)
- [report-inventory.md](./report-inventory.md)
