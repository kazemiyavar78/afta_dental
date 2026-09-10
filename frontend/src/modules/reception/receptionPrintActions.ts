import { message } from 'antd';
import type { Organization } from '@/modules/organization/types';
import { useReceptionStore } from './store/receptionStore';
import { printReceptionReceipt } from './printReceptionReceipt';

/**
 * چاپ قبض پذیرش با وضعیت فعلی استور.
 * @param organizations لیست سازمان‌ها برای نام بیمه
 * @param hasPermission بررسی مجوز خواندن پذیرش
 * @returns true اگر چاپ با موفقیت آغاز شد
 */
export function printReceptionFromStore(
  organizations: Organization[],
  hasPermission: (permission: string) => boolean,
): boolean {
  if (!hasPermission('reception.read')) {
    message.error('شما مجوز این عملیات را ندارید');
    return false;
  }

  const store = useReceptionStore.getState();
  const hasPatient =
    Boolean(store.patient.first_name || store.patient.last_name || store.patient.national_code) ||
    store.patient.id != null;
  if (!hasPatient) {
    message.warning('ابتدا اطلاعات بیمار را تکمیل کنید');
    return false;
  }

  try {
    const insuranceName = organizations.find((o) => o.id === store.insuranceId)?.name ?? '';
    const additionalInsuranceName =
      organizations.find((o) => o.id === store.additionalInsuranceId)?.name ?? '';
    printReceptionReceipt({
      receptionId: store.receptionId,
      receptionDate: store.receptionDate,
      patient: store.patient,
      insuranceName,
      additionalInsuranceName,
      specialCodeValue: store.specialCodeValue,
      specialCodeName: store.specialCodeName,
      referralCode: store.referralCode,
      doctorName: store.doctorName,
      doctorMedicalCode: store.doctorMedicalCode,
      assistantName: store.assistantName,
      description: store.description,
      discount: store.discount,
      services: store.services,
    });
    return true;
  } catch (err: unknown) {
    const msg = err instanceof Error ? err.message : 'خطا در چاپ قبض';
    message.error(msg);
    return false;
  }
}
