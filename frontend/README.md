# فرانت‌اند — سیستم پذیرش بیماران (طب پرداز)

رابط کاربری React با لایه پلتفرم مشترک امنیتی (Session-Cookie + CSRF).

## استک

| مؤلفه | انتخاب |
|---|---|
| React 18 + TypeScript (strict) | ✅ |
| Vite | build/dev |
| React Router v6 | مسیریابی |
| TanStack Query | fetch/cache |
| Axios | HTTP + interceptors |
| Zustand + Context | وضعیت Auth |
| Ant Design | UI + RTL |
| react-hook-form + zod | فرم‌ها |

**قرارداد تایپ:** در کل پروژه از `type` استفاده شده (نه `interface`).

## اجرا

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

فرانت روی `http://localhost:5173` و API از طریق **پروکسی Vite** به `http://localhost:8080` می‌رود.

### پیش‌نیاز بک‌اند

```bash
# در ریشه پروژه
go run ./cmd/server
```

مطمئن شوید `SECURE_COOKIES=false` در `.env` بک‌اند برای توسعه محلی HTTP باشد.

## امنیت

- `withCredentials: true` روی تمام درخواست‌ها
- CSRF: خواندن `csrf_token` از کوکی + هدر `X-CSRF-Token` برای POST/PUT/PATCH/DELETE
- Session در کوکی HttpOnly — **هیچ توکنی در localStorage/sessionStorage ذخیره نمی‌شود**
- 401 → هدایت به `/login` (بدون حلقه بی‌پایان)
- 403 → Toast خطا، بدون ریدایرکت
- `PermissionGuard` — عدم رندر (نه disable) برای عناصر بدون مجوز

## ساختار

```
src/platform/     ← لایه مشترک (api, auth, components, hooks, layout, routes)
src/modules/      ← ماژول‌های کسب‌وکار
```

## افزودن ماژول جدید (چک‌لیست)

1. پوشه `src/modules/نام-ماژول/` بساز با این فایل‌ها:
   - `api.ts` — فراخوانی‌های HTTP با `httpClient`
   - `types.ts` — تایپ‌های داده
   - `hooks.ts` — اسکیمای zod و هوک‌های اختصاصی
   - `pages/` — صفحات لیست و فرم
2. یک آیتم به `src/platform/routes/routeConfig.ts` اضافه کن:
   ```ts
   { path: '/my-module', element: <MyListPage />, requiredPermission: 'my-module.read' },
   ```
3. آیتم منو را در `src/platform/layout/AppShell.tsx` اضافه کن (با `permission`)
4. مجوزهای نقش را در بک‌اند (`internal/user/permissions.go`) به‌روز کن
5. **هیچ کد امنیتی جدید لازم نیست** — از `httpClient`, `PermissionGuard`, `useApiQuery`, `useApiMutation` استفاده کن

## Build تولید

```bash
npm run build
```

خروجی در `dist/` — برای استقرار پشت IIS در کنار بک‌اند روی همان دامنه.
