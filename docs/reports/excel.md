# Excel Export

## Library

`xlsx` (SheetJS)

## Behavior

- Exports **all filtered data** (not current UI page)
- Persian headers from column titles
- Summary row included when present
- Multiple sheets for sectioned reports
- Filename: `{title}-{jalali-date}.xlsx`

## Column Values

- `currency`/`number` → numeric cell
- Custom via `excelFormatter`

## Usage

Enabled via `toolbar.excel: true` in ReportDefinition.
