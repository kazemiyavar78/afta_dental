# Project context — afta_tpdenta

Dental-clinic reception system (`github.com/tpdenta/afta-reception`). Backend and frontend live in one repo. UI copy and comments are Persian. Security design follows AFTA-style controls (session, CSRF, integrity hashes, audit chain, login lockout).

## 1. Purpose and target users

- Clinic staff use a Persian, RTL web app to register patients, open dental receptions, price services against insurance organizations, take payments, and print reports.
- User types in `internal/user/model.go`: `User`, `Doctor`, `Assistant`, `Specialist`.
- Access is permission-based (`internal/user/permissions.go`). A user who holds every catalog permission is treated as admin (`HasAllPermissions`). `cmd/seed-admin` creates an `Admin` role and one admin user.
- Roles named `Reception` and `Doctor` are not created by seed. If those roles already exist and have no permissions, `user.Service.ensureEmptyOperationalRoles` assigns a default set.
- Target users are receptionists, doctors, assistants, and administrators of a dental clinic. No patient-facing portal exists in this repo.

## 2. Tech stack and versions

Versions below are from `go.mod`, `frontend/package.json`, and `frontend/tsconfig.app.json`.

**Backend**

- Go `1.26.4` (`go.mod`). README names the same version.
- HTTP: `github.com/gin-gonic/gin` v1.10.0
- DB driver: `gorm.io/driver/sqlserver` v1.5.4 (indirect), `github.com/microsoft/go-mssqldb` v1.7.2 (indirect)
- ORM: `gorm.io/gorm` v1.25.12
- Env files: `github.com/joho/godotenv` v1.5.1 (indirect)
- Passwords: `golang.org/x/crypto` v0.31.0 (bcrypt in `cmd/seed-admin` and user service)
- IDs: `github.com/google/uuid` v1.6.0
- Encryption: AES-256-GCM in `internal/platform/security/encryption/encryption.go` (key must be 32 bytes)
- Integrity: HMAC-SHA256 in `internal/platform/security/integrity`
- Database product: SQL Server via the driver above. Minimum server version is not enforced in code. README says SQL Server 2019+.

**Frontend** (`frontend/package.json`, private package version `0.0.0`)

- React 19.2.7, React DOM 19.2.7
- TypeScript ~6.0.2, target ES2022 (`frontend/tsconfig.app.json`)
- Vite ^8.1.1 (lock 8.3.6), `@vitejs/plugin-react` ^6.0.3
- Vitest ^5.0.0 (lock 5.1.0), jsdom ^29.1.1, Testing Library React ^16.3.3
- UI: antd ^6.5.0, `@ant-design/icons` ^6.3.2, `antd-jalali-v5` ^1.1.0
- Data: `@tanstack/react-query` ^5.101.2, axios ^1.18.1 (lock 1.21.0), zustand ^5.0.14
- Forms: react-hook-form ^7.81.0, zod ^4.4.3 (lock 4.6.7), `@hookform/resolvers` ^5.4.0
- Routing: react-router-dom ^7.18.1 (lock 7.18.5, and the same for `react-router`; `follow-redirects` lock 1.16.2)
- Dates: dayjs ^1.11.21, jalaliday ^3.1.1
- Export: `@react-pdf/renderer` ^4.9.0, xlsx ^0.18.5
- Node.js `22.17.0` (`.nvmrc` and `frontend/package.json` `engines`). This is the version installed on the development machine and is the Node 22 LTS line. CI installs it via `node-version-file: .nvmrc`.

**Schema changes**

- Startup calls GORM `AutoMigrate` through `internal/platform/migrate.Migrate` (`cmd/server/main.go`). If any model fails, the returned error names that table and `main` exits (FPT_FLS.1). The process does not keep serving with a partial schema.
- schema فقط با GORM AutoMigrate هنگام راه‌اندازی ساخته می‌شود؛ فایل‌های SQL قدیمی هرگز اجرا نشده بودند و حذف شدند (در تاریخچه git موجودند). Q6 روی پایگاه توسعه (۲۰۲۶-۱۱-۰۵) جدول `schema_migrations` را نداشت؛ `Users.UserType` انتهای جدول است و `Receptions` شکل مدل Go را دارد.

## 3. Folder structure

- `cmd/server/main.go` — process entry: config, DB, migrate, security, HTTP, optional static SPA.
- `cmd/seed-admin/main.go` — one-shot admin role and user.
- `internal/platform/config` — env loading.
- `internal/platform/db` — SQL Server connection (max 25 open, 10 idle, 5 min lifetime).
- `internal/platform/migrate` — GORM AutoMigrate / DropTable.
- `internal/platform/testutil` — shared test helpers: Gin context, in-memory audit recorder, and (integration build only) a SQL Server transaction that rolls back.
- `internal/platform/apperror` — API error codes and user messages. `ErrValidation` is code `VALIDATION_ERROR`, HTTP 400, user message «اطلاعات ورودی نامعتبر یا ناقص است.»
- `internal/platform/middleware` — recovery, headers, request log, CORS, rate limit, session, CSRF, security-code check, permission check, and `BindJSON` / `BindQuery`. After the session is found, one username lookup stores `username` in the Gin context. `ActorFromContext` reads user id, username, IP, and User-Agent. If no user is in the context, the id is nil.
- `internal/platform/security/encryption` — AES-GCM and per-entity integrity hashes.
- `internal/platform/security/integrity` — HMAC signer.
- `internal/platform/security/session` — `Sessions` table and `session_id` cookie. Closing a session writes `SESSION_TERMINATED` for the closer and `LOGOUT` for the owner, with the same reason. The full session id is not in the description. The first eight characters can be in details. Reaching the session cap closes the oldest sessions as actor `system` with reason `max_sessions`.
- `internal/platform/security/settings` — key/value security settings.
- `internal/platform/security/audit` — hash-chained `SecurityEvents` and one `AuditChainStates` row (`LastVerifiedID`, `LastVerifiedHash`, `AnchorID`, `AnchorHash`, `LastCheckAt`, `LastResult`, `BrokenAtID`, `IntegrityHash` with `INTEGRITY_HMAC_KEY`). `Record` / `RecordTx` write an `Event` (actor, target, result, reason, masked JSON details) with `HashVersion` 2. `logEvent` is unexported and only the audit package tests use it to keep a `HashVersion` 1 row. `computeRowHash` uses the row version, so version 1 rows still verify. `types.go` is the event and reason catalog (Persian label and severity `critical` / `high` / `normal`). `audit.Diff` returns only changed fields as before/after. `MaskSensitive` replaces values of keys `password`, `pass`, `token`, `otp`, `csrf`, `secret`, `session_id`, and `security_code` with `***`. `MaskTail` keeps the last digits of a national code or phone. `audit.Unavailable` maps a failed write to `apperror.ErrAuditUnavailable` (`AUDIT_UNAVAILABLE`, HTTP 503). Business modules, user, role, session, settings, login, log-read, CSRF, and the retention worker call `Record` or `RecordTx` and do not discard that error. Wallet and saved reception write the event inside the same database transaction. `Record` holds a process mutex and inserts inside a transaction that locks the last row with `UPDLOCK, HOLDLOCK`. `RecordTx` inserts on the caller transaction. `CreatedAt` is UTC truncated to a microsecond before the hash is stored. After migrate, startup records `AUDIT_STARTED` for actor `system`. `SIGINT` / `SIGTERM` records `AUDIT_STOPPED`, then `http.Server.Shutdown` with a 10 second timeout.
- `internal/platform/security/loginguard` — per-IP failed-login lock (`LoginAttempts`). While an IP is locked, `AllowBlockedLoginLog` allows at most one `LOGIN_BLOCKED` audit event per minute for that IP. The login handler and the login middleware share that counter on the same guard.
- `internal/platform/security/logretention` — background log-size and integrity worker.
- `internal/platform/security/passwordpolicy` — password rules from settings.
- `internal/platform/security/uploadguard` — upload window limits.
- `internal/user` — users, roles, permissions, login.
- `internal/patient` — patients and national-code checks.
- `internal/organization` — insurers / organizations.
- `internal/organizationPackage` — tariff packages.
- `internal/services` — dental services and excluded services.
- `internal/tariff` — stored and calculated tariffs (`calc.go`).
- `internal/reception` — receptions, service lines, photos.
- `internal/wallet` — cash fund, patient wallet, bank accounts, credit accounts, ledger.
- `internal/cardreader` — fake POS simulator.
- `internal/fund` — named funds (separate from wallet cash fund).
- `internal/specialcode` — third-party “special code” on a reception.
- `internal/regulation` — service-interval rules and required photos.
- `internal/report` — SQL report queries.
- `internal/logs` — security-event list API. The root `.gitignore` and `frontend/.gitignore` use `/logs/` so only a `logs` directory next to that file is ignored. `*.log` still ignores log files. `internal/logs` and `frontend/src/modules/logs` are source and are tracked (D-10).
- `docs/` — this file, IIS notes, AFTA checklist, report design notes, and `docs/afta/check-output/` for check logs. AFTA working docs: `docs/afta/roadmap.md`, `docs/afta/traceability.md`, `docs/afta/settings.md`, `docs/afta/user-guide.md`, `docs/afta/install-guide.md`, `docs/afta/discovery-phase0-1.md`.
- `scripts/afta-check.ps1` — one check run: vet, unit tests, integration tests, govulncheck, frontend test and build, `npm audit`, read-only SQL in `scripts/db-checks/`.
- `scripts/afta-discovery.ps1` — discovery run. Masking and DSN conversion come from `scripts/lib/redact.ps1` and `scripts/lib/db.ps1`. A null SQL parameter is sent as `DBNull`. A query result is returned as one `DataTable` so PowerShell does not unroll the rows.
- `scripts/db-checks/` — optional read-only `.sql` files. `audit_chain_status.sql` counts `SecurityEvents` by type, counts hash-chain breaks with `LAG`, and reads `AuditChainStates`.
- `scripts/lib/redact.ps1` — masks passwords, connection strings, and session ids in script output. `scripts/lib/db.ps1` converts a DSN for those scripts and does not open a connection by itself.
- `.github/workflows/ci.yml` — GitHub Actions on push and pull_request to `main`. Backend job on `windows-latest` (vet, unit test, govulncheck). Frontend job on `ubuntu-latest` (`npm ci`, test, build, `npm audit`). No secrets and no database variables.
- `.nvmrc` — Node `22.17.0` for local tools and for the frontend CI job.
- `frontend/src/main.tsx` — mounts React, enables Jalali dayjs.
- `frontend/src/App.tsx` — RTL Ant Design, Persian locale, React Query, router. No custom theme object.
- `frontend/src/test/renderWithProviders.tsx` — Vitest render with that same ConfigProvider, a QueryClient that does not retry, and MemoryRouter.
- `frontend/src/test-setup.ts` — Vitest setup (Jalali dayjs and `@testing-library/jest-dom`). `vitest.config.ts` already pointed at this file.
- `frontend/src/platform/` — HTTP client, auth, layout, shared components, report engine.
- `frontend/src/modules/` — feature pages (auth, users, roles, reception, patients, organization, organization-packages, services, fund, wallet, special-codes, regulations, tariff, settings, profile, logs, reports).
- `frontend/vite.config.ts` — dev server port 5173, proxies `/api`.
- `uploads/reception` — reception photo files (path built in `reception.NewService`). Directory is created at upload time.

## 4. Core modules and data flow

**Request path** (`cmd/server/main.go`)

1. `godotenv.Load`, then `config.Load`. Missing `DATABASE_URL` or any of the four security keys aborts startup.
2. `db.Connect`. Then `organization.EnsureCenterPackageColumn` and `migrate.Migrate` for the model list in `main`.
3. `settings.Service.SeedDefaults`, `user.Service.SyncPermissions`, role-hash fix, and one-time organization integrity-hash migrations (`OrgIntegrityHashMigrationV2`, `V3`).
4. Background `logretention.Worker` on `LOG_RETENTION_TICKER`.
5. Gin middleware order: recovery, security headers, request logger, CORS, global rate limit. On `/api`: login rate limit, session cookie, role name, permissions, CSRF, user `SecurityCode` check, route permission check.
6. Handlers call a service, which uses a GORM repository and refreshes HMAC / AES integrity fields. Mutating services call `audit.Recorder.Record` or `RecordTx` with `ActorFromContext`. A failed write returns `AUDIT_UNAVAILABLE`. Wallet charges and a saved reception record the event on the open transaction.
7. `DEV_MODE` or `APP_ENV=development`: API only. Otherwise `registerFrontendRoutes` serves `frontend/dist` and falls back to `index.html` for non-API paths.

**Auth**

- `POST /api/login` checks password (bcrypt), lockout, password policy, work hours, and blocked weekdays (`security_check` and user service). A failed attempt writes `LOGIN_FAILED` with reason `user_not_found`, `wrong_password`, `account_locked`, `account_inactive`, `outside_work_hours`, or `integrity_violation`, and the typed username as the target. Only a missing user or a wrong password counts toward the IP lock. Reaching the threshold writes `IP_LOCKED` with `lock_minutes`. An attempt while that IP is locked writes at most one `LOGIN_BLOCKED` per minute (`ip_locked`), both in `LoginRateLimiterMiddleware` and in `HandleUserLogin`. Success writes `SESSION_CREATED` with IP and browser and does not write `LOGIN_SUCCESS`. The HttpOnly cookie is `session_id` (`session.SessionCookieName`). Locking the user account itself is still phase 4.
- Later GET/HEAD/OPTIONS set non-HttpOnly cookie `csrf_token`. Mutating calls must send the same value in `X-CSRF-Token` (`frontend/src/platform/api/httpClient.ts`, `withCredentials: true`).
- Public paths that skip session and CSRF: `/api/login`, `/api/health` (`middleware/session.go`).
- `AuthorizationMiddleware` allows the request if the user is admin or has any one of the route’s `RequirePermission` names.
- Failed `SecurityCode` check deletes that user’s sessions.

**Reception pricing**

- UI (`frontend/src/modules/reception`) loads patient, organizations, services, doctors, assistants, special codes, regulations.
- `POST /api/reception/calculate` prices lines without saving (`tariff` + organization package shares).
- `POST /api/reception` and `PUT /api/reception/:id` upsert patient data inside the reception payload. `save: true` stores status `saved`; otherwise status stays `draft` (`reception/model.go`).
- `POST /api/reception/:id/end` checks regulations and uploaded photos (`EndReceptionResponse`).
- Wallet charges (`/api/wallet/*`) adjust `patient_wallets` and append `wallet_transactions`. Service add/remove also moves `cash_funds`. Credit charges reduce `credit_accounts.used_amount`. Marking a credit transaction paid (`credit_paid`) does not change the limit or wallet balance (comment on `Transaction.CreditPaid` and README behavior matches `SetCreditTransactionPaid`).

**Reports**

- Backend returns JSON rows, summary, and meta (`internal/report`).
- Frontend report engine (`frontend/src/platform/reports`) renders, prints, and exports Excel/PDF in the browser. There is no server export endpoint.

**Frontend data**

- Pages call module `api.ts` through `httpClient`.
- Server state: React Query. Reception workspace state: Zustand (`receptionStore.ts`).
- Jalali dates via dayjs `jalaliday` and `antd-jalali-v5`. Layout is RTL.

## 5. API and UI routes

Base path `/api`. Success bodies are the DTO named below. List endpoints return a JSON array unless noted. All routes except login and health require the session cookie. POST/PUT/PATCH/DELETE also require CSRF.

Permission names are OR. Admin (`HasAllPermissions`) bypasses them.

### Errors

Responses from `middleware.WriteError` are JSON `{ "error": string, "code": string }`. An `apperror.AppError` uses its HTTP status and Persian user message. Any other error is HTTP 500, code `INTERNAL`, message «خطای داخلی سرور رخ داده است.» The original error text is not returned.

Invalid JSON or query input is a Gin error, not an `AppError`. `middleware.BindJSON` and `middleware.BindQuery` answer that with HTTP 400, code `VALIDATION_ERROR`, message «اطلاعات ورودی نامعتبر یا ناقص است.» (`apperror.ErrValidation`, D-08). The response has no field name and no submitted value. The server log records the field name and the validation tag only. URI binding is unchanged.

`UpdateSettingByID` does not write the setting name or value to stdout. If the save fails, the log line is the key name and the error, not the new value (D-09). `CheckWorkHours` does not print the work-hour start or end values. `cmd/seed-admin` still prints the admin username and password; that stays until phase 6.

### Health and auth (`internal/user/routes.go`, `cmd/server/main.go`)

| Method | Path | Permission | Input | Output |
|---|---|---|---|---|
| GET | `/api/health` | public | — | `{ "status": "ok" }` |
| POST | `/api/login` | public | `LoginRequest` username, password | `LoginResponse` user + `session_id`; sets `session_id` cookie |
| POST | `/api/logout` | session | — | `{ "message": "خروج موفق" }` |
| GET | `/api/me` | session | — | `MeResponse` user, permissions, `is_admin` |
| POST | `/api/change-password` | session | `ChangePasswordRequest` old_password, new_password | message |
| GET | `/api/profile` | session | — | `UserProfileResponse` user + sessions |
| GET | `/api/users` | `users.read` | — | `[]UserResponse` |
| POST | `/api/users` | `users.create` | `CreateUserRequest` | `UserResponse` 201 |
| GET | `/api/users/doctors` | `reception.read` or `users.read` | — | doctors (`UserType` Doctor or Specialist) |
| GET | `/api/users/assistants` | `reception.read` or `users.read` | — | assistants |
| GET | `/api/users/:id` | handler: admin or `users.read`, else only self | id | `UserResponse` |
| PUT | `/api/users/:id` | handler: admin or `users.update`, else only self | `UpdateUserRequest` | `UserResponse` |
| GET | `/api/roles/assignable` | `users.create` or `users.update` | — | `[]RoleResponse` |
| GET | `/api/roles` | `roles.read` | — | role list |
| POST | `/api/roles` | `roles.create` | `CreateRoleRequest` name, description, permission_ids | `RoleDetailResponse` 201 |
| GET | `/api/roles/:id` | `roles.read` | id | `RoleDetailResponse` including `integrity_ok` |
| PUT | `/api/roles/:id` | `roles.update` | `UpdateRoleRequest` | `RoleDetailResponse` |
| DELETE | `/api/roles/:id` | `roles.delete` | id | message |
| GET | `/api/permissions` | `roles.read` | — | `[]PermissionResponse` |
| GET | `/api/security/settings` | `security.settings` | — | settings with Persian labels (`SettingsDicFN`) |
| PUT | `/api/security/settings` | `security.settings` | `UpdateSecuritySettingRequest` id, name, value | message |
| GET | `/api/sessions` | handler: admin or `users.listSessions` sees others; optional query `user_id` | — | `[]SessionResponse` |
| DELETE | `/api/sessions/:id` | session | id | message |

### Patients (`internal/patient`)

| Method | Path | Permission | Input | Output |
|---|---|---|---|---|
| GET | `/api/patients` | `patient.read` | `SearchRequest` query filters; empty filter lists all | `[]Response` |
| GET | `/api/patients/last-file-number` | `patient.read` or `reception.read` or `reception.create` | — | `LastFileNumberResponse` |
| POST | `/api/patients` | `patient.create` | `CreateRequest` | `Response` 201 |
| GET | `/api/patients/:id` | `patient.read` | id | `Response` |
| PUT | `/api/patients/:id` | `patient.update` | `UpdateRequest` | `Response` |
| DELETE | `/api/patients/:id` | `patient.delete` | id | soft delete |

`sex`: true = male, false = female (`patient/model.go`). `is_foreign_national` switches national-code rules (`patient/nationalcode.go`).

### Organizations and packages

| Method | Path | Permission | Input | Output |
|---|---|---|---|---|
| GET/POST | `/api/organization` | `organization.read` / `create` | `CreateRequest` name, is_takmili, is_active, is_free, package_id, center_package_id | `Response` or list |
| GET/PUT/DELETE | `/api/organization/:id` | read / update / delete | `UpdateRequest` | `Response` or soft delete |
| GET/POST | `/api/organization-packages` | `organization_packages.read` / `create` | `CreateRequest` coefficients and percentages | package `Response` |
| GET/PUT/DELETE | `/api/organization-packages/:id` | read / update / delete | `UpdateRequest` | package `Response` |

### Services and tariffs

| Method | Path | Permission | Input | Output |
|---|---|---|---|---|
| GET/POST | `/api/services` | `services.read` / `create` | `CreateRequest` | `Response` |
| GET/PUT/DELETE | `/api/services/:id` | read / update / delete | `UpdateRequest` | `Response` |
| GET | `/api/excluded-services/organization/:organizationId` | `excluded_services.read` | org id | `[]ExcludedServiceResponse` |
| POST | `/api/excluded-services` | `excluded_services.add` | `ExcludedServicesRequest` organization_id, service_ids | created rows |
| DELETE | `/api/excluded-services` | `excluded_services.remove` | same JSON body | soft delete |
| POST | `/api/tariff/calculate` | `tariff.read` | `CalculateTariffForOrganizationRequest` | `CalculateResponse` |
| POST | `/api/tariff/save` | `tariff.create` | same body as calculate | `SaveTariffResponse` |
| GET | `/api/tariff/organization/:organizationId` | `tariff.read` | org id | tariff rows |
| PUT | `/api/tariff/:id/recalculate` | `tariff.update` | `RecalculateTariffRequest` three center amounts | `TariffResponse` |

`tariff.delete` exists in the permission catalog. No delete route is registered.

### Reception (`internal/reception/routes.go`)

| Method | Path | Permission | Input | Output |
|---|---|---|---|---|
| GET | `/api/reception` | `reception.read` | — | reception list |
| GET | `/api/reception/nav` | `reception.read` | `NavQuery` cursor, dir | `ReceptionResponse` (empty flag if none) |
| GET | `/api/reception/next-file-number` | `reception.read` or `create` | — | next file number |
| POST | `/api/reception/calculate` | `reception.create` or `update` | `CalculateRequest` | `CalculateResponse` |
| GET | `/api/reception/patient/:patientId` | `reception.read` | patient id | receptions |
| GET | `/api/reception/patient/:patientId/services-history` | `reception.read` | patient id | `[]PatientServiceHistoryItem` |
| POST | `/api/reception` | `reception.create` | `UpsertReceptionRequest` | `ReceptionResponse` 201 |
| GET | `/api/reception/:id` | `reception.read` | id | `ReceptionResponse` |
| PUT | `/api/reception/:id` | `reception.update` | `UpsertReceptionRequest` | `ReceptionResponse` |
| DELETE | `/api/reception/:id` | `reception.delete` | id | soft delete |
| POST | `/api/reception/:id/restore` | `reception.restore` | id | restored reception |
| POST | `/api/reception/:id/end` | `reception.update` | id | `EndReceptionResponse` |
| POST | `/api/reception/:id/photos` | `reception.update` | multipart field `file` | `ReceptionPhotoResponse` 201 |
| GET | `/api/reception/:id/photos/:photoId` | `reception.read` | ids | file bytes |

### Wallet, banks, credit, card reader (`internal/wallet`, `internal/cardreader`)

Amounts on credit accounts are rial (`CreditAccount` comments). Other money fields are `int64` without a unit comment, except `Funds.Balance` which is `float64`.

| Method | Path | Permission | Input | Output |
|---|---|---|---|---|
| GET | `/api/wallet/cash-fund` | `wallet.read` | — | `CashFundResponse` |
| GET | `/api/wallet/:fileId/balance` | `wallet.read` | file id (patient id) | `WalletBalanceResponse` |
| GET | `/api/wallet/:fileId/transactions` | `wallet.read` | file id | `[]TransactionResponse` |
| POST | `/api/wallet/cash` | `wallet.cash` | `CashAdjustRequest` | `TransactionResponse` 201 |
| POST | `/api/wallet/card-to-card` | `wallet.card_to_card` | `CardToCardAdjustRequest` | `TransactionResponse` 201 |
| POST | `/api/wallet/card-reader` | `wallet.card_reader` | `CardReaderChargeRequest` (approved must be true to charge) | `TransactionResponse` 201 |
| POST | `/api/wallet/credit-charge` | `credit_charge.create` | `CreditChargeRequest` | `TransactionResponse` 201 |
| PATCH | `/api/wallet/transactions/:id/credit-paid` | `credit_charge.update` | `SetCreditPaidRequest` | `TransactionResponse` |
| GET | `/api/credit-accounts` | `credit_account.read` or `reports.read` | — | `[]CreditAccountResponse` |
| GET | `/api/credit-accounts/active` | `credit_charge.create` or `credit_account.read` | — | active accounts |
| POST | `/api/credit-accounts` | `credit_account.create` | `CreateCreditAccountRequest` | `CreditAccountResponse` 201 |
| GET/PUT | `/api/credit-accounts/:id` | read / update | `UpdateCreditAccountRequest` | `CreditAccountResponse` |
| DELETE | `/api/credit-accounts/:id` | `credit_account.delete` | id | 200 only if `used_amount` is 0 |
| GET/POST | `/api/bank-accounts` | `bank_account.read` / `create` | `CreateBankAccountRequest` | `BankAccountResponse` |
| GET/PUT/DELETE | `/api/bank-accounts/:id` | read / update / delete | `UpdateBankAccountRequest` | delete rejected if transactions exist |
| POST | `/api/card-reader/simulate` | `wallet.card_reader` | `ChargeRequest` amount | `ChargeResponse` fake POS data, always approved |

### Funds, special codes, regulations

| Method | Path | Permission | Input | Output |
|---|---|---|---|---|
| GET/POST | `/api/fund` | `fund.read` / `fund.create` | `CreateRequest` name | `Response` id, name, balance |
| GET | `/api/fund/:id` | `fund.read` | id | `Response` |
| GET/POST | `/api/special-codes` | `special_code.read` / `create` | `CreateRequest` | `Response` |
| GET | `/api/special-codes/by-code/:code` | `special_code.read` | code | `Response` |
| GET/PUT/DELETE | `/api/special-codes/:id` | read / update / delete | `UpdateRequest` | `Response` |
| GET/POST | `/api/regulations` | `regulation.read` / `create` | `CreateRequest` service_ids, duration_days, photo_count, description, is_active | `Response` |
| GET/PUT/DELETE | `/api/regulations/:id` | read / update / delete | `UpdateRequest` | `Response` |

`fund.update` and `fund.delete` are in the permission catalog. No update or delete routes exist.

### Logs (`internal/logs`)

| Method | Path | Permission | Input | Output |
|---|---|---|---|---|
| GET | `/api/logs` | `logs.read` | query ip, username, from_date, to_date (`2006-01-02`), optional `page` | `[]SecurityEvent`. `page` empty or `1` writes `AUDIT_READ` with those filters and no row bodies. A later page does not. Denial of this path writes `AUDIT_READ_DENIED`. |
| GET | `/api/logs/user` | session only | optional `page` | events for the current user. First page writes `AUDIT_READ` with `scope=self`. |

### Reports (`internal/report/routes.go`)

All require `reports.read`. Query binds to the struct named below. Response is that report’s `*Response` (rows, summary, meta; some include `sections`).

- `GET /api/reports/revenue-by-organization` — `RevenueByOrganizationQuery` (from_date, to_date required).
- `GET /api/reports/doctor-performance/by-organization` — `DoctorPerformanceByOrganizationQuery` (doctor_ids required).
- `GET /api/reports/doctor-performance/by-service` — `DoctorPerformanceByServiceQuery` (doctor_ids and service_ids required).
- `GET /api/reports/doctor-performance/patient-list` — `DoctorPerformancePatientListQuery`.
- `GET /api/reports/doctor-performance/revenue-by-doctors` — `RevenueByDoctorsQuery`.
- `GET /api/reports/user-performance/by-reception` — `UserPerformanceByReceptionQuery` (user_ids required).
- `GET /api/reports/user-performance/by-transaction` — `UserPerformanceByTransactionQuery`.
- `GET /api/reports/credit-account-payments` — `CreditAccountPaymentsQuery` (optional `credit_account_id`).

ID list query params are comma-separated strings (`doctor_ids`, `user_ids`, `service_ids`, organization id lists).

### Frontend routes (`frontend/src/platform/routes/routeConfig.tsx`)

`/` and unknown paths redirect to `/users`. Each route’s `requiredPermission` matches the page’s main read permission. `/login` is public. `/profile` has no permission flag.

`/login`, `/users`, `/users/new`, `/users/:id/edit`, `/users/:id/sessions`, `/roles`, `/reception`, `/reception/new`, `/patients`, `/organization`, `/organization/new`, `/organization/:id/edit`, `/organization-packages`, `/services`, `/fund`, `/fund/new`, `/bank-accounts`, `/credit-accounts`, `/special-codes`, `/regulations`, `/tariff`, `/tariff/new`, `/settings`, `/logs`, `/reports/revenue-by-organization`, `/reports/doctor-performance`, `/reports/user-performance`, `/reports/credit-account-payments`, `/profile`.

## 6. Data models

GORM models. `gorm.Model` adds `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt` (soft delete) unless a model declares its own primary key.

| Model | Table | Notes |
|---|---|---|
| `user.User` | `Users` | Username unique. PasswordHash bcrypt. SecurityCode is AES-wrapped SHA-256 of name, family, phone, password hash, address, role, medical code. IntegrityHash is HMAC. UserType: User, Doctor, Assistant, Specialist. TwoFactorEnabled / TwoFactorSecret columns exist. |
| `user.Role` | `Roles` | Name unique, IntegrityHash. |
| `user.Permission` | `Permissions` | Name unique. Catalog is `defaultPermissions`. |
| `user.RolePermission` | `RolePermissions` | Composite key RoleID + PermissionID, IntegrityHash. |
| `session.Session` | `Sessions` | UUID `Id` (SQL Server uniqueidentifier), Ip, Browser, CreationTime, PersonnelAccountID. |
| `settings.SecuritySetting` | `SecuritySettings` | SettingName unique, SettingValue, IntegrityHash. `LastVerifiedEventId` is not a setting; startup deletes that row if it still exists. |
| `audit.SecurityEvent` | `SecurityEvents` | Hash chain PrevHash / RowHash. `CreatedAt` is microsecond UTC. `UserID` is the actor id. `ActorUsername` is that username at event time (not a join). Nullable: `ActorUsername`, `TargetType`, `TargetID`, `TargetName`, `Result` (`success`/`failure`), `Reason`, `UserAgent`, `Details` (masked JSON). `HashVersion` tinyint NOT NULL default 1. Version 2 hashes the new fields too. |
| `audit.AuditChainState` | `AuditChainStates` | Single row. Checkpoint and cleanup anchor for the hash chain. `IntegrityHash` is HMAC with `INTEGRITY_HMAC_KEY`, not the audit chain key. |
| `loginguard.LoginAttempt` | `LoginAttempts` | PK Ip, FailedCount, LockedUntil. |
| `logretention.LogCleanupSummary` | `LogCleanupSummaries` | Written by the background worker. Included in startup `Migrate`. |
| `uploadguard.UploadCounter` | `UploadCounters` | Atomic per-user upload window counter. Included in startup `Migrate`. |
| `patient.Patient` | `patients` | Unique national_code and file_number. IntegrityHash is AES-wrapped hash of identity fields. Fields themselves are stored in plaintext. |
| `organization.Organization` | `Organizations` | Name, IsTakmili, IsActive, IsFree, PackageID, CenterPackageID, IntegrityHash. |
| `organizationpackage.OrganizationPackage` | `OrganizationPackages` | Coefficients and subsidy / supplementary / organization percentages. IntegrityHash. |
| `organizationpackage.OrganizationPackageRelation` | `OrganizationPackageRelations` | OrganizationID + PackageID. AutoMigrated. No repository methods use it. Organizations point at packages with PackageID and CenterPackageID. |
| `services.ServiceItem` | `Services` | Coefficients, rates, tariffs, specialist rate/tariff (excluded from integrity hash per comment), tooth flags. |
| `services.ExcludedService` | `ExcludedServices` | Unique (OrganizationID, ServiceID), HMAC. |
| `tariff.Tariff` | `Tariffs` | Per organization and service: amount, tariff, organization / supplementary / subsidy shares. |
| `reception.Reception` | `Receptions` | Patient, optional base and supplementary insurance, special code, doctor, assistant, dates, status `draft` or `saved`, discount, coverage, ReceptionEnded. |
| `reception.ReceptionService` | `ReceptionServices` | Priced line, optional tooth number and direction 1–4. |
| `reception.ReceptionPhoto` | `ReceptionPhotos` | FilePath, FileName. |
| `wallet.CashFund` | `cash_funds` | Single running clinic-income balance. |
| `wallet.PatientWallet` | `patient_wallets` | One row per FileID. |
| `wallet.BankAccount` | `bank_accounts` | Bank, sheba, account, card, name, HMAC. |
| `wallet.CreditAccount` | `credit_accounts` | Name, description, credit_limit, used_amount, is_active, HMAC. Amounts documented as rial. |
| `wallet.Transaction` | `wallet_transactions` | Category `cash_fund` or `wallet`. Actions: service_added, service_removed, charge, payment, refund. Methods: cash, card_reader, card_to_card, check, wallet_credit, credit_account. CreditPaid does not change balances. |
| `fund.Fund` | `Funds` | ID, Name, Balance float64. No update/delete API. |
| `specialcode.SpecialCode` | `SpecialCodes` | Unique Code, percentage 0–100, HMAC. |
| `regulation.Regulation` | `Regulations` | DurationDays, PhotoCount, IsActive, Description. |
| `regulation.RegulationService` | `RegulationServices` | RegulationID + ServiceID. |

Integrity for wallet, bank, credit, special code, and excluded service is HMAC (`integrity.Signer`). Integrity for user security code, patient, organization, package, and service is SHA-256 of concatenated fields, then AES-GCM.

## 7. Environment variables

Names and purpose only. No values.

| Name | Where | Purpose |
|---|---|---|
| `DATABASE_URL` | backend, required | SQL Server DSN for GORM. |
| `AES_KEY` | backend, required | 32-byte AES-256-GCM key. |
| `INTEGRITY_HMAC_KEY` | backend, required | HMAC key for row integrity. Separate from AES. |
| `AUDIT_HMAC_KEY` | backend, required | HMAC key for the security-event hash chain. |
| `CSRF_HMAC_KEY` | backend, required | Key used when issuing CSRF cookie tokens. |
| `SERVER_HOST` | backend | Bind host. Code default `0.0.0.0`. |
| `SERVER_PORT` | backend | Bind port. Code default `8080`. |
| `APP_ENV` | backend | `development` turns on dev mode (API only, Gin debug). Other values, including the code default `production`, serve `frontend/dist`. |
| `DEV_MODE` | backend | `true` forces dev mode. |
| `SECURE_COOKIES` | backend | Secure flag on session and CSRF cookies. Code default true. |
| `ALLOWED_ORIGINS` | backend | Comma-separated CORS origins. |
| `RATE_LIMIT_PER_MINUTE` | backend | Global request cap. Code default 120. |
| `LOGIN_RATE_LIMIT_PER_MINUTE` | backend | Loaded in config. Login limiting also uses `loginguard` and `MaximumNumberOfFailedLogin`. UNKNOWN: this specific variable is not read outside `config.Load`. |
| `LOG_RETENTION_TICKER` | backend | Interval for log cleanup and integrity worker. Go duration. Code default 15m. |
| `INTEGRITY_TICKER` | backend | Loaded into `Config.IntegrityTicker`. Not passed to the worker. The worker uses `LOG_RETENTION_TICKER` only. |
| `SEED_ADMIN_USERNAME` | `cmd/seed-admin` | Username for the seeded admin. Code fallback exists. |
| `SEED_ADMIN_PASSWORD` | `cmd/seed-admin` | Password for the seeded admin. Code fallback exists. Do not commit it. |
| `AFTA_TEST_DATABASE_URL` | `go test -tags integration`, `scripts/afta-check.ps1` | SQL Server DSN for integration tests. Not read by the server. If empty, those tests call `t.Skip`. The check script may ask once to reuse `DATABASE_URL` from `.env` for a development database. It refuses a production-like name and the system databases `master`, `model`, `msdb`, and `tempdb`. |
| `VITE_API_BASE_URL` | frontend | Axios base URL. Example file sets `/api`. |
| `VITE_API_PROXY_TARGET` | `frontend/vite.config.ts` | Dev-server proxy target for `/api`. Code fallback `http://127.0.0.1:8080`. Not listed in `frontend/.env.example`. |

Root `.env.example` is referenced by README and is not in the tree. `frontend/.env.example` only documents `VITE_API_BASE_URL`. A root `.env` file exists and is not documented here.

Security setting keys (database, not env) are in `internal/platform/security/settings/keys.go`: session cap, failed logins, IP lock minutes, log table size, work hours, per-weekday blocked flags, password rules, upload caps, and two organization hash-migration flags. The old `LastVerifiedEventId` key is removed from defaults and from the Persian labels. Startup deletes a leftover row and writes `SETTING_REMOVED`. Chain progress lives in `AuditChainStates`.

## 8. Commands

From the repo root unless noted. There is no Makefile or Dockerfile.

**Backend**

- Dependencies: `go mod download`
- Run API: `go run ./cmd/server` (loads `.env` if present)
- Seed admin: `go run ./cmd/seed-admin` (needs the same required env as the server; permissions must already exist, so run the server once first — seed calls `FindAllPermissions` and fails if that query fails)
- Unit test (no SQL Server): `go test ./...`
- Vet: `go vet ./...`
- Integration test (SQL Server, skipped when `AFTA_TEST_DATABASE_URL` is empty): `go test -tags integration ./...`
- Build: `go build -o server.exe ./cmd/server`
- Full check (the operator runs this; it writes `docs/afta/check-output/check-<yyyyMMdd-HHmm>.txt`): `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\afta-check.ps1`
- CI (no database, no secrets): `.github/workflows/ci.yml` on push and pull_request to `main`. Backend (`windows-latest`): `actions/setup-go` with `go-version-file: go.mod` and module cache, then `go vet ./...`, `go test ./... -count=1` (no `integration` tag), `govulncheck ./...`. Frontend (`ubuntu-latest`): Node from `.nvmrc`, npm cache, `npm ci`, `npm test`, `npm run build`, `npm audit --omit=dev --audit-level=high` in `frontend/`. `govulncheck` and `npm audit` use `continue-on-error` until phase 5.

Database create script in README: `CREATE DATABASE afta_reception`. The running app uses whatever database is in `DATABASE_URL`.

**Frontend** (directory `frontend/`)

- Install: `npm install`
- Dev (port 5173, proxies `/api`): `npm run dev`
- Test: `npm test` (`vitest run`). Shared render helper: `src/test/renderWithProviders.tsx`.
- Typecheck and production bundle: `npm run build` (`tsc && vite build`)
- Preview the bundle: `npm run preview`

Dev workflow: backend with `DEV_MODE=true` or `APP_ENV=development`, plus `npm run dev`. Production: `npm run build`, then run the Go server with dev mode off so it serves `frontend/dist`.

## 9. Architectural decisions visible in the code

- Modular `internal/<feature>` packages with handler, service, repository, and DTO. `main` wires them by hand. No generic DI container.
- Schema is only GORM AutoMigrate at startup (`internal/platform/migrate`). A failed table is named in the error and stops the process (FPT_FLS.1). The old SQL files under `migrations/` were never applied (no `schema_migrations` on the development database) and were removed; they remain in git history.
- SQL Server-specific SQL: `MERGE` for login attempts, `sys.dm_db_partition_stats` for log size, `CONVERT`/`CAST` in patient search, custom `MSSQLUUID` so uniqueidentifier bytes match the driver.
- Separate HMAC and AES keys. User/patient/org/service integrity stores an encrypted hash (`SecurityCode` / `IntegrityHash`). Wallet-style rows store a raw HMAC and keep older field lists so rows written before credit-account columns still verify (`VerifyTransactionIntegrity`).
- Permission strings are code-owned and upserted on startup (`SyncPermissions`). Admin is “has every permission”, not a hardcoded role check on every route. The `Admin` role is granted every permission when it exists.
- Double-submit CSRF plus HttpOnly session cookie. SameSite is Strict. Dev uses the Vite proxy so the browser sees one origin.
- Reports stay read-only JSON on the server. Print, Excel, and PDF run in the browser (`frontend/src/platform/reports`) so the server does not render documents.
- Card reader is an in-process simulator (`internal/cardreader`). The package comment says there is no device connection.
- Money for credit accounts is integer rial. Shares on a tariff line are stored integers (organization, supplementary, subsidy) rather than recomputed on every read.
- One `RoleID` per user. Comments in README say a future `UserRoles` table could be added. That table does not exist.
- Background worker verifies the audit hash chain, then trims only `SecurityEvents` rows whose ID is less than or equal to `LastVerifiedID`, using `MaximumSizeOfLogTables` as a megabyte cap. The last deleted row's hash is stored as the chain anchor. A successful check updates `AuditChainStates` and does not write a success audit event. `DATA_TAMPERING` is written only when `BrokenAtID` changes. The worker also re-checks user and role integrity.

## 10. Known issues, gaps, and TODOs

- D-01 is done pending the consultant. `docs/afta/check-output/discovery-20261105-2123.txt` (section Q6) lists `LogCleanupSummaries` and `UploadCounters` on the development database. `TestIntegration_AutoMigrate_AllModelsCreate` and `TestIntegration_LogCleanupAndUploadCounterTablesExist` remain integration tests and are not part of `go test ./...`.
- R1-02, R1-04, R1-06, R1-07, R1-08, R1-17, R1-21, and PP-01 are done pending the consultant. Role, user, session, settings, and login events use `Record` with the request actor. A failed write returns `AUDIT_UNAVAILABLE`. Session close is two events and the description has no full session id. Settings changes store the Persian label and the value before and after. User updates store only changed fields; password hash and security code stay out of that diff. Log read is recorded on the first page only, and a denied `/api/logs` request records `AUDIT_READ_DENIED`. The log screen still does not show the new columns (R1-01, D-16).
- R1-18 is in progress. Failed login, IP lock duration, and at most one blocked-login event per minute per IP are recorded. Locking the user account itself remains phase 4 (R1-15).
- D-13 is waiting for the database check. Patient, organization, package, service, special code, regulation, fund, and bank account create events are covered by unit tests. Reception draft create and tariff save are `//go:build integration`. CSRF and log retention also call `Record`. A failed write returns `AUDIT_UNAVAILABLE`. Startup records `AUDIT_STARTED` and shutdown records `AUDIT_STOPPED`. `TestIntegration_SecurityEventNewColumnsExist`, `TestIntegration_D13_ReceptionDraftRecordsActorAndTarget`, and `TestIntegration_D13_TariffSaveRecordsActorAndTarget` are waiting for the database check.
- D-14 is done pending the consultant. CSRF calls `Record` with `ActorFromContext`. When no user is in the context, `UserID` stays nil. If that write fails, the response is `AUDIT_UNAVAILABLE` (HTTP 503) instead of 403.
- D-11, D-12, D-23, and R1-24 are waiting for the database check. Chain verification reads forward from `AuditChainStates`, cleanup keeps an anchor hash, and `LastVerifiedEventId` is no longer a security setting. The 3861 `DATA_TAMPERING` rows in the development database were false alarms from the old checkpoint bug. The evaluation install uses a clean database. Unit tests cover the chain without SQL Server. `TestIntegration_D11_ConcurrentLogAndVerifyChain` is `//go:build integration` and is not part of `go test ./...`. `scripts/db-checks/audit_chain_status.sql` is read-only.
- D-31 is done. High advisories for axios, follow-redirects, react-router/react-router-dom, and zod were closed by `npm audit fix` without `--force`. The lockfile pins axios 1.21.0, follow-redirects 1.16.2, react-router 7.18.5, react-router-dom 7.18.5, and zod 4.6.7. The `package.json` ranges already allowed those versions, so only `frontend/package-lock.json` changed. `npm audit --omit=dev` then reports only `xlsx@0.18.5`. That package has no npm fix and stays until phase 5 (PP-18). Until then, `govulncheck` and `npm audit` in `.github/workflows/ci.yml` stay `continue-on-error` and must become blocking in phase 5 (PP-19).
- Root `.env.example` is not in the repository. README lists environment variable names and purpose only. `frontend/.env.example` documents `VITE_API_BASE_URL`.
- File headers in `internal/organization`, `internal/fund`, and matching frontend organization/fund files still say the module is a sample for later phases. Those packages have working CRUD (fund is create/read only).
- `OrganizationPackageRelation` is migrated and unused. Organization rows store `PackageID` and `CenterPackageID` directly.
- `fund.update`, `fund.delete`, and `tariff.delete` permissions have no routes.
- `User.TwoFactorEnabled` and `TwoFactorSecret` are stored and included in the user HMAC (`user/repo.go`). No login or settings handler reads or sets them.
- `INTEGRITY_TICKER` is loaded and unused. `LOGIN_RATE_LIMIT_PER_MINUTE` is loaded and not referenced again.
- Column names are mixed and are left as they are. Fields from `gorm.Model` use the default naming strategy (`id`, `created_at`, `updated_at`, `deleted_at`, snake_case). Fields with an explicit `column:` tag stay PascalCase (for example `UserType`, `Balance`). Renaming them would need a data migration; this pass does not change them.
- `fund.Fund.Balance` is `float64` and the development column type is float. Backend rules forbid float for money. It is not changed in this pass because existing rows need a data migration. Tracked as D-02 for phase 5.
- `HasTransactions` on bank-account create/update responses is hardcoded `false` (`wallet/service.go` around the create and update mappings). List and get compute it from transactions.
- Card payments are simulated. `SimulateCharge` always returns `Approved: true`.
- `frontend/src/platform/components/DataTable/DataTable.tsx`: `TODO: pagination سمت سرور — فعلاً کلاینت‌ساید`.
- `frontend/src/modules/users/hooks.ts`: `TODO: خواندن قوانین پیچیدگی رمز از endpoint تنظیمات امنیتی بک‌اند`. Password rules are enforced on the server (`passwordpolicy`) and not loaded into that form.
- `docs/reports/migration-report.md` lists open report work: bundle Vazirmatn locally, lazy-load the PDF exporter, a backend export endpoint for large datasets, a print adapter for the reception receipt, and code-splitting `@react-pdf/renderer`. The same file notes PDF font from CDN, a 70mm reception receipt outside the tabular report engine, and a large PDF bundle.
- `cmd/server/main.go` contains commented `DropTable` calls. They are not executed.
- Seed prints the admin username and password to stdout (`cmd/seed-admin/main.go`).
- Patient, organization, and service business fields are stored in plaintext. AES protects the integrity token, not the columns.
- Doctor and Reception roles are not created by seed. Default permissions apply only when a role with that exact name already exists and has zero permissions.

## 11. Tests

Unit tests must not open SQL Server. `go test ./...` does not pass `-tags integration`, so files with `//go:build integration` are left out.

Integration tests use that build tag and `AFTA_TEST_DATABASE_URL`. `internal/platform/testutil.OpenTestDB` skips the test when the variable is empty. `WithTestTx` runs the test body in a transaction and rolls it back, including when the test fails.

Shared Go helpers live in `internal/platform/testutil`:

- `NewTestGinContext` builds an `httptest` request and a Gin context. `SetUserID`, `SetUsername`, `SetRoleName`, `SetPermissions`, and `SetSession` write the same context keys as `internal/platform/middleware` (`userID`, `username`, `roleName`, `permissions`, `session`, `sessionID`).
- `FakeAuditRecorder` stores `Record` and `RecordTx` in memory, including the full `audit.Event`. It implements `audit.Recorder`. Session and settings tests cannot import `testutil` (that package imports them) and use a local in-memory recorder. User tests and the business-module create tests use `FakeAuditRecorder`. `TestR1_02_RoleCreatedRecordsActorAndTarget`, `TestR1_04_RoleUpdatedListsPermissionDiff`, `TestR1_06_MaxSessionsRecordsSystemLogout`, `TestR1_07_DeleteSessionRecordsActorTargetAndReason`, `TestR1_17_SettingChangedRecordsBeforeAndAfter`, `TestR1_18_LoginFailedReasons`, `TestR1_18_LoginBlockedOncePerMinute`, `TestR1_21_UserUpdatedDiffOmitsPasswordAndUnchanged`, `TestFAU_GEN1_AuditReadOnlyFirstPage`, `TestFAU_GEN1_AuditReadDenied`, `TestD13_PatientCreatedRecordsActorAndTarget`, `TestD13_FundCreatedRecordsActorAndTarget`, `TestD13_RegulationCreatedRecordsActorAndTarget`, `TestD13_SpecialCodeCreatedRecordsActorAndTarget`, `TestD13_OrganizationCreatedRecordsActorAndTarget`, `TestD13_OrganizationPackageCreatedRecordsActorAndTarget`, `TestD13_ServiceCreatedRecordsActorAndTarget`, and `TestD13_BankAccountCreatedRecordsActorAndTarget` cover the new events without SQL Server.

`TestWriteError_AppErrorReturnsCodeAndMessage` and `TestWriteError_UnknownErrorIsGeneric500` call `middleware.WriteError` through that Gin helper. The unknown-error case expects HTTP 500, the public internal message, and no leak of the original error text. `TestD08_BindJSON_EmptyOrIncompleteReturnsValidationError` expects HTTP 400 and `VALIDATION_ERROR` for an empty body and for a missing required field, with no field name in the response. `TestD08_BindJSON_ValidBodyFillsStruct` expects `BindJSON` to return true and fill the struct. `TestD08_BindQuery_InvalidValueDoesNotLeakFieldValue` expects the submitted query value to stay out of the response and the server log. `TestComputeRowHash_V1MatchesManualHMAC` and `TestComputeRowHash_V2IncludesNewFields` compare both hash versions with an independent HMAC. `TestHashVersions_MixedChainStillVerifies` keeps a version 1 row valid beside a version 2 row. `TestRecord_FullActor`, `TestRecord_SystemActor`, and `TestMaskSensitive_ReplacesSecretKeys` cover the new writer. `TestD14_ActorFromContextWithoutUserIsNil` uses `testutil` and expects a nil user id. `TestErrAuditUnavailable_StatusAndMessage` locks the 503 audit failure. No injectable clock was added; `computeRowHash` already takes the timestamp. `TestD11_HealthyChainAdvancesState`, `TestD11_TamperedDescriptionStopsAtPreviousRow`, `TestD11_RepeatCheckDoesNotRepeatTamperEvent`, `TestD12_CleanupPrefixThenVerifyRemaining`, `TestD11_ConcurrentLogEventKeepsSingleChain`, and `TestCreatedAt_MicrosecondRoundTripKeepsHash` use an in-memory audit repository. `TestR1_24_RemoveRetiredCheckpointAuditsOnce` checks that a leftover `LastVerifiedEventId` row is deleted once and audited as `SETTING_REMOVED`.

`TestIntegration_AutoMigrate_AllModelsCreate` runs `migrate.Migrate` on the same model list as `cmd/server/main.go` and fails if any table returns an error. `TestIntegration_LogCleanupAndUploadCounterTablesExist` then reads `INFORMATION_SCHEMA.TABLES` for `LogCleanupSummaries` and `UploadCounters`. `TestIntegration_D11_ConcurrentLogAndVerifyChain` writes 100 events from two goroutines and expects `VerifyChain` from an empty checkpoint to stay intact. `TestIntegration_SecurityEventNewColumnsExist` checks the new `SecurityEvents` columns after `Migrate`. None of these are part of `go test ./...`. `TestCombineMigrateErrors_EmptyIsNil` and `TestCombineMigrateErrors_NamesEveryTable` check error aggregation without a database.

Frontend component tests use `renderWithProviders` (RTL, `fa_IR`, default Ant Design theme, React Query with retry off, MemoryRouter). `PageHeader` has a sample test. Vitest setup stays `frontend/src/test-setup.ts`. `frontend/d31-production-deps.test.ts` reads `package-lock.json` and expects the D-31 pins (axios 1.21.0, follow-redirects 1.16.2, react-router and react-router-dom 7.18.5, zod 4.6.7) and unchanged `xlsx@0.18.5`. It lives outside `src` so `tsc` does not need Node types.

`scripts/afta-check.ps1` runs, in order, `GO_VET`, `GO_TEST_UNIT` (`go test ./... -count=1`), `GO_TEST_INTEGRATION` (`go test -tags integration ./... -count=1`), `GOVULNCHECK` (prints `go install golang.org/x/vuln/cmd/govulncheck@latest` and skips when the command is missing), `FRONT_TEST`, `FRONT_BUILD`, `NPM_AUDIT` (`npm audit --omit=dev` in `frontend/`), and `DB_CHECKS` (every `scripts/db-checks/*.sql`, read-only, each in a transaction that rolls back, including `audit_chain_status.sql`). A failed section does not stop the later sections. The text file ends with one status line per section. Output goes through `scripts/lib/redact.ps1`. Add new database checks as `.sql` files in that folder; do not add a second check script.

`.github/workflows/ci.yml` repeats the no-database parts on GitHub: vet, `go test ./... -count=1`, non-blocking `govulncheck`, frontend test, frontend build, and non-blocking `npm audit --omit=dev --audit-level=high`. It does not run integration tests or `scripts/db-checks/`. After D-31 that production audit reports only `xlsx@0.18.5`. `govulncheck` and `npm audit` stay non-blocking until phase 5, when that package is replaced and both steps must fail the job.
