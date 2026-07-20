import { useEffect, useRef, useState } from 'react';
import type { KeyboardEvent } from 'react';
import { Button, Input, InputNumber, Select, Space, message } from 'antd';
import type { InputRef } from 'antd/es/input';
import { DeleteOutlined } from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { fetchServices } from '@/modules/services/api';
import type { ServiceLineState } from '../types';
import { useReceptionStore } from '../store/receptionStore';
import { ToothChartModal } from './ToothChartModal';

/** گزینه‌های جهت دندان (چهار ربع) */
const DIRECTION_OPTIONS = [
  { value: 1, label: 'راست بالا' },
  { value: 2, label: 'چپ بالا' },
  { value: 3, label: 'راست پایین' },
  { value: 4, label: 'چپ پایین' },
];

type ServiceRowProps = {
  line: ServiceLineState;
  editing: boolean;
  /** سهم صندوق این سطر (نرخ − سازمان − تکمیلی − یارانه) */
  cashAmount: number;
  /** true یعنی این آخرین سطر جدول است؛ Enter/Tab روی توضیحات → خدمت جدید */
  isLastRow: boolean;
  /** کلید سطر بعدی؛ برای فوکوس تعداد وقتی ردیف پایین وجود دارد */
  nextLineKey: string | null;
  /** قبل از انتخاب خدمت، وجود سازمان را بررسی می‌کند؛ false یعنی مسدود */
  ensureOrganization: () => boolean;
  onChange: (patch: Partial<ServiceLineState>) => void;
  onRemove: () => void;
  onRecalculate: () => void;
  /**
   * افزودن سطر خدمت جدید.
   * فقط از روی توضیحات آخرین سطر (Enter/Tab) یا از بیرون (دکمه / Ctrl+Space) فراخوانی می‌شود.
   */
  onAddNextLine: () => void;
};

/** فوکوس به اولین input داخل ظرف (Select / InputNumber آنت دیزاین) */
function focusField(wrap: HTMLElement | null) {
  wrap?.querySelector<HTMLElement>('input')?.focus();
}

/**
 * یک سطر جدول خدمات با پشتیبانی جهت و دندان و فوکوس ترتیبی.
 * زنجیره فوکوس داخل سطر: خدمت → تعداد → جهت/دندان → توضیحات
 */
export function ServiceRow({
  line,
  editing,
  cashAmount,
  isLastRow,
  nextLineKey,
  ensureOrganization,
  onChange,
  onRemove,
  onRecalculate,
  onAddNextLine,
}: ServiceRowProps) {
  const [toothOpen, setToothOpen] = useState(false);

  // توکن‌های فوکوس سراسری از استور (افزودن سطر / پرش از توضیحات سطر بالا)
  const serviceFocusKey = useReceptionStore((s) => s.serviceFocusKey);
  const quantityFocusKey = useReceptionStore((s) => s.quantityFocusKey);
  const clearServiceFocus = useReceptionStore((s) => s.clearServiceFocus);
  const clearQuantityFocus = useReceptionStore((s) => s.clearQuantityFocus);
  const requestQuantityFocus = useReceptionStore((s) => s.requestQuantityFocus);

  const serviceWrapRef = useRef<HTMLDivElement>(null);
  const quantityWrapRef = useRef<HTMLDivElement>(null);
  const directionWrapRef = useRef<HTMLDivElement>(null);
  const toothWrapRef = useRef<HTMLDivElement>(null);
  const descRef = useRef<InputRef>(null);

  const { data: servicesData } = useQuery({
    queryKey: ['services'],
    queryFn: fetchServices,
  });
  const services = servicesData ?? [];

  // پس از «افزودن خدمت»، فوکوس روی Select خدمت همین سطر
  useEffect(() => {
    if (serviceFocusKey === line.key && editing) {
      window.setTimeout(() => {
        focusField(serviceWrapRef.current);
        clearServiceFocus();
      }, 0);
    }
  }, [serviceFocusKey, line.key, editing, clearServiceFocus]);

  // وقتی از توضیحات سطر بالایی Enter/Tab زده شود و این سطر وجود داشته باشد → فوکوس تعداد
  useEffect(() => {
    if (quantityFocusKey === line.key && editing) {
      window.setTimeout(() => {
        focusField(quantityWrapRef.current);
        clearQuantityFocus();
      }, 0);
    }
  }, [quantityFocusKey, line.key, editing, clearQuantityFocus]);

  /**
   * پس از تأیید تعداد (Enter): اگر خدمت جهت دارد → جهت،
   * وگرنه اگر دندان دارد → شماره دندان، وگرنه → توضیحات.
   */
  function focusAfterQuantity(nextLine: Partial<ServiceLineState> = {}) {
    const hasDir = nextLine.has_dental_direction ?? line.has_dental_direction;
    const hasTooth = nextLine.has_tooth ?? line.has_tooth;
    window.setTimeout(() => {
      if (hasDir) {
        focusField(directionWrapRef.current);
      } else if (hasTooth) {
        focusField(toothWrapRef.current);
      } else {
        descRef.current?.focus();
      }
    }, 0);
  }

  /** انتخاب خدمت از لیست، پر کردن پرچم‌های جهت/دندان، محاسبه مجدد، فوکوس تعداد */
  function pickService(serviceId: number) {
    if (!ensureOrganization()) return;
    const svc = services.find((s) => s.id === serviceId);
    if (!svc) return;
    onChange({
      service_id: svc.id,
      service_code: svc.service_code,
      service_name: svc.name,
      quantity: svc.default_count > 0 ? svc.default_count : 1,
      has_dental_direction: svc.is_dental_direction,
      has_tooth: svc.has_tooth,
      teeth_number: null,
      teeth_direction: null,
    });
    onRecalculate();
    window.setTimeout(() => focusField(quantityWrapRef.current), 0);
  }

  /**
   * Enter یا Tab روی فیلد توضیحات:
   * - اگر ردیف پایین‌تری وجود داشته باشد → فوکوس تعداد همان ردیف
   * - اگر آخرین ردیف باشد → افزودن خدمت جدید
   * Shift+Tab عمداً دست‌نخورده می‌ماند تا برگشت عادی کار کند.
   */
  function handleDescriptionKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (!editing) return;
    const isAdvance = e.key === 'Enter' || (e.key === 'Tab' && !e.shiftKey);
    if (!isAdvance) return;

    e.preventDefault();
    e.stopPropagation();

    if (!isLastRow && nextLineKey) {
      // ردیف بعدی از قبل هست → فقط برو روی تعداد آن
      requestQuantityFocus(nextLineKey);
      return;
    }

    // آخرین سطر → خدمت جدید بساز (استور خودش فوکوس Select خدمت را تنظیم می‌کند)
    onAddNextLine();
  }

  return (
    <>
      <tr className={cashAmount < 0 ? 'reception-service-negative' : undefined}>
        <td className="col-service">
          <div ref={serviceWrapRef}>
            <Select
              showSearch
              optionFilterProp="label"
              style={{ minWidth: 160, width: '100%' }}
              disabled={!editing}
              value={line.service_id || undefined}
              options={services.map((s) => ({
                value: s.id,
                label: `${s.service_code} — ${s.name}`,
              }))}
              onChange={pickService}
              onFocus={() => {
                if (!ensureOrganization()) {
                  message.warning(
                    'لطفاً قبل از انتخاب خدمت، یک سازمان (پایه یا تکمیلی) انتخاب کنید.',
                  );
                }
              }}
            />
          </div>
        </td>
        <td className="col-qty">
          <div ref={quantityWrapRef}>
            <InputNumber
              min={1}
              disabled={!editing || !line.service_id}
              value={line.quantity}
              onChange={(v) => {
                onChange({ quantity: Number(v) || 1 });
                onRecalculate();
              }}
              onPressEnter={() => focusAfterQuantity()}
            />
          </div>
        </td>
        <td className="col-direction">
          {line.has_dental_direction ? (
            <div ref={directionWrapRef}>
              <Select
                allowClear
                style={{ minWidth: 110, width: '100%' }}
                disabled={!editing || !line.has_dental_direction}
                value={line.teeth_direction ?? undefined}
                options={DIRECTION_OPTIONS}
                onChange={(v) => {
                  onChange({ teeth_direction: v ?? null });
                  // بعد از جهت: اگر دندان لازم است → دندان، وگرنه توضیحات
                  window.setTimeout(() => {
                    if (line.has_tooth) focusField(toothWrapRef.current);
                    else descRef.current?.focus();
                  }, 0);
                }}
              />
            </div>
          ) : (
            <span className="cell-muted">—</span>
          )}
        </td>
        <td className="col-tooth">
          {line.has_tooth ? (
            <div ref={toothWrapRef}>
              <Space.Compact style={{ width: '100%' }}>
                <InputNumber
                  min={1}
                  max={8}
                  disabled={!editing || !line.has_tooth}
                  value={line.teeth_number ?? undefined}
                  placeholder={line.has_tooth ? '۱–۸' : undefined}
                  style={{ width: '100%' }}
                  onKeyDown={(e) => {
                    // Space Alone → نمودار دندان؛ Ctrl+Space برای افزودن خدمت در سطح جدول است
                    if (e.code === 'Space' && !e.ctrlKey && line.has_tooth && editing) {
                      e.preventDefault();
                      setToothOpen(true);
                    }
                  }}
                  onChange={(v) => onChange({ teeth_number: v == null ? null : Number(v) })}
                  onPressEnter={() => descRef.current?.focus()}
                />
                {line.has_tooth && editing && (
                  <Button onClick={() => setToothOpen(true)} title="انتخاب از نمودار دندان">
                    نمودار
                  </Button>
                )}
              </Space.Compact>
            </div>
          ) : (
            <span className="cell-muted">—</span>
          )}
        </td>
        <td className="col-amount num">{Number(line.service_amount ?? 0).toLocaleString('fa-IR')}</td>
        <td className="col-tariff num">{Number(line.service_tariff ?? 0).toLocaleString('fa-IR')}</td>
        <td className="col-org num">
          {Number(line.service_organization_share ?? 0).toLocaleString('fa-IR')}
        </td>
        <td className="col-supp num">
          {Number(line.service_supplementary_insurance_share ?? 0).toLocaleString('fa-IR')}
        </td>
        <td className="col-subsidy num">
          {Number(line.service_subsidy_share ?? 0).toLocaleString('fa-IR')}
        </td>
        <td className="col-cash num">{Number(cashAmount ?? 0).toLocaleString('fa-IR')}</td>
        <td className="col-desc">
          <Input
            ref={descRef}
            disabled={!editing}
            value={line.service_description}
            onChange={(e) => onChange({ service_description: e.target.value })}
            onKeyDown={handleDescriptionKeyDown}
          />
        </td>
        <td>
          {editing && (
            <Button danger type="text" icon={<DeleteOutlined />} onClick={onRemove} />
          )}
        </td>
      </tr>
      <ToothChartModal
        open={toothOpen}
        selectedNumber={line.teeth_number}
        selectedDirection={line.teeth_direction}
        onClose={() => setToothOpen(false)}
        onSelect={({ teeth_number, teeth_direction }) =>
          onChange({ teeth_number, teeth_direction })
        }
      />
    </>
  );
}
