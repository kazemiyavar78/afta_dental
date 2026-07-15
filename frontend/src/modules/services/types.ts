export type ServiceFeatures = '' | '#' | '*' | '#*';

export type ServiceItem = {
  id: number;
  service_code: string;
  name: string;
  technical_coefficient: number;
  professional_coefficient: number;
  consumption_coefficient: number;
  service_rate: number;
  service_tariff: number;
  international_code: string;
  default_count: number;
  maximum_count: number;
  service_features: ServiceFeatures;
  is_active: boolean;
  is_dental_direction: boolean;
  allow_multiple_use: boolean;
};

export type ServicePayload = {
  service_code: string;
  name: string;
  technical_coefficient: number;
  professional_coefficient: number;
  consumption_coefficient: number;
  service_rate: number;
  service_tariff: number;
  international_code: string;
  default_count: number;
  maximum_count: number;
  service_features: ServiceFeatures;
  is_active: boolean;
  is_dental_direction: boolean;
  allow_multiple_use: boolean;
};
