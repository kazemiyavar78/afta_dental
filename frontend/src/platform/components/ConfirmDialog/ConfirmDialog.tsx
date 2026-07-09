import { Modal } from 'antd';
import type { ModalFuncProps } from 'antd';

type ConfirmOptions = ModalFuncProps & {
  onConfirm: () => void | Promise<void>;
};

/** دیالوگ تأیید استاندارد */
export function confirmDialog({ onConfirm, title, content, ...rest }: ConfirmOptions) {
  Modal.confirm({
    title: title ?? 'تأیید عملیات',
    content,
    okText: 'بله',
    cancelText: 'انصراف',
    onOk: onConfirm,
    ...rest,
  });
}
