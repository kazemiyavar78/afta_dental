import { Component, type ErrorInfo, type ReactNode } from 'react';
import { Result, Button } from 'antd';

type Props = { children: ReactNode };
type State = { hasError: boolean };

/** مرز خطا برای جلوگیری از کرش کل اپ */
export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false };

  static getDerivedStateFromError(): State {
    return { hasError: true };
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    console.error('ErrorBoundary:', error, info);
  }

  render() {
    if (this.state.hasError) {
      return (
        <Result
          status="error"
          title="خطای غیرمنتظره"
          subTitle="مشکلی در نمایش این بخش رخ داده است."
          extra={
            <Button type="primary" onClick={() => window.location.reload()}>
              بارگذاری مجدد
            </Button>
          }
        />
      );
    }
    return this.props.children;
  }
}
