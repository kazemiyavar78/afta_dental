export type Patient = {
  id: number;
  first_name: string;
  last_name: string;
  national_code: string;
  birth_date: string;
  address: string | null;
  home_phone_number: string | null;
  mobile_phone_number: string | null;
  file_number: string;
  /** true = مرد ، false = زن */
  sex: boolean;
};

export type PatientPayload = {
  first_name: string;
  last_name: string;
  national_code: string;
  birth_date: string;
  address?: string | null;
  home_phone_number?: string | null;
  mobile_phone_number?: string | null;
  file_number: string;
  sex: boolean;
};

export type PatientSearchParams = {
  first_name?: string;
  last_name?: string;
  national_code?: string;
  birth_date?: string;
  address?: string;
  home_phone_number?: string;
  mobile_phone_number?: string;
  file_number?: string;
  sex?: boolean;
};
