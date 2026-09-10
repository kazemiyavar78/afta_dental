# PDF Export

## Library

`@react-pdf/renderer` — real PDF rendering (not HTML screenshot).

## Persian / RTL

- Font: Vazirmatn (CDN registration)
- Direction: RTL on Page
- All text through `pdfFormatter` or default `formatColumnValue`

## Usage

PDF export is triggered via `ReportToolbar` → handled by `useReport` → `exportReportPdf`.

## Limitations

| Case | Limitation | Workaround |
|---|---|---|
| Reception Receipt 70mm | Custom thermal layout | Keep custom HTML print |
| Very large reports | Browser memory | Backend export endpoint (future) |
| Offline PDF | CDN font required | Bundle Vazirmatn in `public/fonts/` |

## Adding PDF to New Report

No extra code needed if `ReportDefinition` is complete. Toolbar `pdf: true` enables it.
