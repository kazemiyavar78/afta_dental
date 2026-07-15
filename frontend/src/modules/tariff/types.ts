import type { ServiceItem } from '@/modules/services/types';
import type { Organization } from '@/modules/organization/types';

/** مبالغ محاسبه‌شده برای یک خدمت */
export type TariffCalculate = {
  total_amount: number;
  tariff: number;
  organization_share: number;
  supplement_amount: number;
  subsidy_amount: number;
  fund_amount: number;
};

/** خدمت همراه با نتیجه محاسبه */
export type ServiceWithPrice = {
  service: ServiceItem;
  calculate: TariffCalculate;
};

/** بدنه تست/ذخیره تعرفه */
export type CalculateTariffPayload = {
  organization_id: number;
  exclude_service_ids: number[];
  technical_amount: number;
  professional_center_amount: number;
  consumption_center_amount: number;
};

/** پاسخ تست محاسبه تعرفه */
export type CalculateTariffResponse = {
  services: ServiceWithPrice[];
  organization: Organization;
  send_info: CalculateTariffPayload;
};

/** ردیف تعرفه ذخیره‌شده */
export type Tariff = {
  id: number;
  organization_id: number;
  service_id: number;
  service_code: string;
  service_name: string;
  amount: number;
  tariff_amount: number;
  organization_share: number;
  supplementary_share: number;
  subsidy_share: number;
  fund_amount: number;
};

/** پاسخ ذخیره گروهی */
export type SaveTariffResponse = {
  items: Tariff[];
};

/** بدنه بازمحاسبه یک تعرفه */
export type RecalculateTariffPayload = {
  technical_amount: number;
  professional_center_amount: number;
  consumption_center_amount: number;
};
