// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

export type Reception = {
  id: number;
  patient_name: string;
  doctor_id: number;
  reception_date: string;
  status: string;
};

export type CreateReceptionPayload = {
  patient_name: string;
  doctor_id: number;
  reception_date: string;
};
