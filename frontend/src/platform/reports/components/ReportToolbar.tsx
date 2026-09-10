import { Button, Space } from 'antd';
import {
  FileExcelOutlined,
  FilePdfOutlined,
  PrinterOutlined,
  ReloadOutlined,
  SettingOutlined,
  ClearOutlined,
} from '@ant-design/icons';
import type { ReportToolbarOptions } from '../types';

type ReportToolbarProps = {
  options?: ReportToolbarOptions;
  onPrint?: () => void;
  onPdf?: () => void;
  onExcel?: () => void;
  onPageSetup?: () => void;
  onRefresh?: () => void;
  onResetFilters?: () => void;
  disabled?: boolean;
  loading?: boolean;
};

/**
 * Toolbar مشترک گزارش — Print, PDF, Excel, Page Setup.
 */
export function ReportToolbar({
  options,
  onPrint,
  onPdf,
  onExcel,
  onPageSetup,
  onRefresh,
  onResetFilters,
  disabled,
  loading,
}: ReportToolbarProps) {
  const opts = {
    print: true,
    pdf: true,
    excel: true,
    pageSetup: true,
    refresh: false,
    resetFilters: false,
    ...options,
  };

  return (
    <Space wrap>
      {opts.print && (
        <Button icon={<PrinterOutlined />} onClick={onPrint} disabled={disabled} loading={loading}>
          چاپ
        </Button>
      )}
      {opts.pdf && (
        <Button icon={<FilePdfOutlined />} onClick={onPdf} disabled={disabled} loading={loading}>
          PDF
        </Button>
      )}
      {opts.excel && (
        <Button icon={<FileExcelOutlined />} onClick={onExcel} disabled={disabled} loading={loading}>
          Excel
        </Button>
      )}
      {opts.pageSetup && (
        <Button icon={<SettingOutlined />} onClick={onPageSetup} disabled={disabled}>
          تنظیمات صفحه
        </Button>
      )}
      {opts.refresh && onRefresh && (
        <Button icon={<ReloadOutlined />} onClick={onRefresh}>
          بروزرسانی
        </Button>
      )}
      {opts.resetFilters && onResetFilters && (
        <Button icon={<ClearOutlined />} onClick={onResetFilters}>
          پاک کردن فیلترها
        </Button>
      )}
    </Space>
  );
}
