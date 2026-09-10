# Report Engine Architecture

## Overview

```
modules/<name>/
    api.ts          ← HTTP (httpClient)
    types.ts        ← Business types
    definitions/    ← ReportDefinition (Single Source of Truth)
    pages/          ← Filters + Query + ReportViewer

platform/reports/
    components/     ← ReportViewer, ReportTable, ReportToolbar, ...
    exporters/      ← print.ts, pdf.tsx, excel.ts
    formatters/     ← number, date, value
    hooks/          ← useReport
    types.ts        ← Core types
```

## Single Source of Truth

```
                 ReportDefinition
                        │
        ┌───────────────┼────────────────┐
        │               │                │
        ▼               ▼                ▼
    Ant Design        PDF              Excel
      Table             │                │
        │               │                │
        └───────────────┴────────────────┘
                        │
                      Print
```

## Separation of Concerns

| Layer | Responsibility |
|---|---|
| Module | Filters, API, business logic, ReportDefinition |
| Engine | Presentation, formatting, export, page config |
| Platform | httpClient, auth, shared UI |

## Runtime Page Config

`ReportPageSetup` تنظیمات size/orientation/margin را به‌صورت runtime مدیریت می‌کند بدون تغییر `ReportDefinition` اصلی.

## Sections

گزارش‌های با تفکیک (مثل پزشکان جدا) از `ReportSection[]` پشتیبانی می‌کنند.

## Special Documents

قبض پذیرش 70mm یک **Document Print** است نه tabular report — layout custom دارد ولی از formatters مشترک استفاده می‌کند.
