# Report Definition

## Example

```ts
import type { ReportDefinition } from '@/platform/reports';
import { formatCurrency } from '@/platform/reports';

export const myReport: ReportDefinition<MyRow> = {
  id: 'my-report',
  title: 'گزارش نمونه',
  page: { size: 'A4', orientation: 'portrait', margin: 12 },
  columns: [
    { key: 'name', title: 'نام', align: 'right', format: 'text' },
    { key: 'amount', title: 'مبلغ', format: 'currency', formatter: (v) => formatCurrency(v) },
  ],
  summary: {
    label: 'جمع کل',
    fields: [
      { key: 'amount', label: 'جمع', format: 'currency' },
    ],
  },
  toolbar: { print: true, pdf: true, excel: true, pageSetup: true },
};
```

## Usage

```tsx
const { data, isLoading } = useMyReport(filters);

return (
  <ReportViewer
    definition={myReport}
    data={data?.rows ?? []}
    summary={data?.summary}
    meta={buildMetaRows(data?.meta)}
    loading={isLoading}
    rowKey="id"
  />
);
```

## Column Options

| Field | Description |
|---|---|
| `key` | Field key in row data |
| `title` | Persian header |
| `format` | text, number, currency, date, datetime, percent |
| `formatter` | UI render (ReactNode) |
| `printFormatter` | Print-specific string |
| `pdfFormatter` | PDF-specific string |
| `excelFormatter` | Excel raw value |
| `hidden` | Hide from UI |
| `printable: false` | Exclude from print |
| `exportable: false` | Exclude from Excel |

## Custom Formatter

```ts
{
  key: 'status',
  title: 'وضعیت',
  formatter: (value) => <Tag>{value}</Tag>,
  printFormatter: (value) => String(value),
  excelFormatter: (value) => String(value),
}
```
