# Print Export

## Method

HTML/CSS → `window.open` → `window.print()`

Print does **not** generate PDF first.

## CSS Features

- Dynamic `@page { size; margin; }`
- `@media print`
- RTL direction
- `thead { display: table-header-group }` for repeating headers
- `page-break-inside: avoid` on rows

## Page Setup

User can change size/orientation/margin via `ReportPageSetup` modal — applied as runtime config.

## Custom Documents

For non-tabular prints (reception receipt), use shared utilities:

```ts
import { escapeHtml, formatCurrency } from '@/platform/reports';
```

Build custom HTML but reuse formatters.
