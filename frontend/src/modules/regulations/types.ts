/** انواع داده ماژول ضوابط */
export type Regulation = {
  id: number;
  service_ids: number[];
  duration_days: number;
  is_active: boolean;
  photo_count: number;
  description: string;
};

export type RegulationPayload = {
  service_ids: number[];
  duration_days: number;
  is_active: boolean;
  photo_count: number;
  description: string;
};
