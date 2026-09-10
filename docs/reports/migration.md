# Migration Guide

## Steps per Report

```
Inventory → Definition → ReportViewer → Verify → Cleanup
```

## Checklist

- [ ] Create `ReportDefinition` in `definitions/`
- [ ] Move column formatters to definition
- [ ] Move summary config to definition
- [ ] Build meta rows helper for filters
- [ ] Replace DataTable + print button with `ReportViewer`
- [ ] Verify Print output matches previous
- [ ] Verify PDF/Excel (new capability)
- [ ] Delete old print file
- [ ] Update inventory status

## Adapter Pattern

```ts
type ReportDataAdapter<TInput, TOutput> = (data: TInput) => TOutput[];

const rows = adapter(apiResponse);
```

## Pages with Action Columns

If table has CRUD action columns (e.g. Patients), keep `DataTable` for UI and use `ReportToolbar` + `useReport` for export only.

## Migrated Reports

| Report | Status |
|---|---|
| Revenue by Organization | ✅ Migrated |
| Doctor Performance (3 types) | ✅ Migrated |
| Patient List Print | ✅ Migrated (toolbar only) |
| Reception Receipt | ⚠️ Partial — shared formatters only |
