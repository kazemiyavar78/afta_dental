/** انواع داده ماژول کد خاص */
export type SpecialCode = {
  id: number;
  code: string;
  name: string;
  description: string;
  percentage: number;
  is_active: boolean;
};

export type SpecialCodePayload = {
  code: string;
  name: string;
  description: string;
  percentage: number;
  is_active: boolean;
};
