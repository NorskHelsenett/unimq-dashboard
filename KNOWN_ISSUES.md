# Known Issues

Findings from a read-only review of the repository. This is a backlog/triage document.

Focus was the **Go backend**; frontend, Helm and dev-environment findings are included where noticed.

Severity: **Critical** (data loss / security breach / core feature broken) · **High** (feature silently wrong) · **Medium** (incorrect behaviour, poor DX, scaling risk) · **Low** (cosmetic, cleanup, docs).

Existing items already tracked in `todos.md` (pagination, alarm cleanup, unique receivers, retry logic, connectivity watchers, more unit tests, Dex redirect to backend) are **not** repeated here except where a concrete defect was found.

Fixed issues have been moved out of the backlog into the **FIXED — CHANGE LOG** tables at the bottom. The tables in the body are the outstanding work only.
Remaining rows are marked ⚠️ Partial/Regressed or ❌ Not fixed where a fix was attempted; unmarked rows have not been addressed. Original descriptions are left unchanged for context.

---

## RE-REVIEW STATUS (baseline `7efe67b` → `6fc2fe6`, tenth pass)

`go build ./...`, `go vet ./...` and `go test ./...` all pass.

**G20, G21, G31, G32, G34 and G35 are all fixed.** The duplicate `/rabbitmq/usage` route is gone, the vhost-usage annotation matches the real path, `checkMaintenanceSchedules` uses `break` plus an `IsConfigured()` guard so the entry is always marked notified, all seven handlers now read `{vhost-name}`/`{queue-id}`, and the `sendEmail` guard is the right way round.

⚠️ **G34 and G35 were both Critical regressions that existed on `main` for two commits and were invisible to `go build`, `go vet` and the whole test suite.** A renamed route parameter and a missing `!` each took out a subsystem. Neither is covered by a test *now* either — see **O1**; a table-driven route test and a two-case email-guard test would pin both down permanently.

**No new issues found this pass.** The remaining backlog is unchanged.

*Earlier context:* the seventh pass covered a large restructuring — Prometheus removed end-to-end, RabbitMQ handlers split into `internal/api/v1/rmq/`, routes moved to `internal/api/routes/`, `httpsuite` moved to `internal/api/httpsuite/`, notifications split into `email.go`/`webhook.go` with per-recipient status reporting, and Swagger given an OAuth2 authorization-code + PKCE flow.

### 🚨 Still open — fix these first

| # | What happens now | Where | Impact | Fix |
|---|------------------|-------|--------|-----|
| **G22** | **`MaintenanceEntry` cannot parse its own JSON output.** Marshalling uses the default `time.Time` format (RFC3339), but the custom `UnmarshalJSON` parses via `timehelper.ParseTimeInUTC`, whose layout is `2006-01-02 15:04:05`. | `internal/models/maintenance.go:108-150`; `internal/helpers/timehelper/timehelper.go:5` | **Medium.** Verified: marshal → unmarshal of a valid entry fails with `invalid start time format: parsing time "2026-09-24T06:50:07Z" as "2006-01-02 15:04:05"`. Any round-trip (client replay, fixture, cached copy) is rejected. | Add a matching `MarshalJSON`, or accept RFC3339 on input. |
| **G23** | **`updated_at` is required but can never be omitted.** `UnmarshalJSON` rejects an empty `updated_at`, while the struct tag is `omitempty` on a `time.Time` — and `omitempty` has **no effect on structs**. | `internal/models/maintenance.go:103,133-135` | **Medium.** Verified: a never-edited entry serialises as `"updated_at":"0001-01-01T00:00:00Z"`, so the API advertises a zero timestamp as though it were real. The required-check masks it rather than fixing it. | Use `*time.Time` (or a string) so `omitempty` works, and drop the required-check for an audit field the client shouldn't supply. |
| **G33** | **No guard against the committed Swagger spec drifting.** The `Checker` re-tag in `7efe67b` shipped without regenerating `internal/docs/` (you have since regenerated it — I verified `swag init` now produces **zero diff**). | `internal/docs/{docs.go,swagger.json,swagger.yaml}`; `.github/workflows/testandbuild_backend.yaml` | **Low.** Generated artefacts are committed, so they drift silently whenever `swag init` is skipped, and nothing catches it. CI regenerates the docs but never compares them against the committed copy. | Add a CI step that runs `swag init` and fails if `git diff --exit-code internal/docs` is non-empty. |

### ⚠️ Worth a second look

- **G24 — swallowed error in `sendEmail`.** `mail.NewClient` failures set `status.Error` but **do not `return`**, so execution continues and the real cause is overwritten by the subsequent `DialAndSend` error. Verified: an empty `EMAIL_SMTP_HOST` reports `dial tcp :25: connection refused` instead of `hostname for client cannot be empty`. `internal/helpers/notificationhelper/email.go:54-57`.
- **G25 — status errors serialise to `{}`.** `EmailStatus.Error` and `WebhookStatus.Error` are typed `error` with a `json:"error"` tag. Verified: both marshal to `"error":{}`, so the per-recipient detail the refactor added is **lost the moment it reaches the client**. Store a `string` alongside, or implement `MarshalJSON`. `email.go:19`, `webhook.go:16`.
- **G26 — dead error branches in the test-notification handler.** After `notificationStatus.WebhookStatuses = SendWebhooks(...)` the code checks `if err != nil`, but `err` is a **stale** variable left over from the earlier `GetVhost` call (provably `nil` there). The same pattern repeats after `SendEmails`, which makes the `errors.Is(err, ErrEmailNotConfigured)` warning unreachable. `internal/api/v1/notificationrule.go:538,549`.
- **G27 — `castSliceToStringSlice` is dead.** Defined in `internal/api/httpsuite/context.go:107`, never called.
- **G28 — `ValidationErrors` is now dead.** `IsRequestValid` returns the raw `validator` error, so nothing constructs a `ValidationErrors` any more — but the type, its `Error()` method and the doc comment promising "returns a ValidationErrors instance" all remain. `internal/api/httpsuite/validation.go:7-34`.
- **G29 — `models/prometheus.go` survived the Prometheus removal.** `Sample` and `RangeOptions` have no remaining references.
- **G30 — maintenance is marked notified regardless of delivery.** *Improved but not closed.* `SetMaintenanceEntryNotified` now runs **after** sending rather than before, and every path reaches it (the G31 `return` is gone), but it is still called unconditionally — so an entry whose webhooks and emails all failed is still recorded as notified and never retried. `internal/notify/checker.go:344`.
- **B14 leftover** — `e.Notified = false` in `UnmarshalJSON` is immediately overwritten two lines later. Harmless, but delete it.
- **`time, err := ...`** in `MaintenanceEntry.UnmarshalJSON` shadows the `time` package. Legal, but rename it.
- **A2** — unchanged. `IsRequestValid` is still called from exactly one handler (`notificationrecipient.go:131`), and `validate:` tags exist only in `models/notification.go`.
- **D2** — still partial; sentinels exist but the 404 mapping is not applied uniformly.
- **R7, S11** — unchanged.
- **`internal/database/interface.go`** — unchanged; the `Store[T]` multi-instantiation constraint still applies.

---

## BACKEND

### Correctness / logic bugs

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| B16 | `CheckVhostExists` decodes notification docs into `[]AlarmEntry` | `internal/database/vhost.go:37-60` | Wrong target type (copy-paste). Works only because it just checks `len() > 0`. Fragile and misleading. | **Medium** | Use `CountDocuments` with the `_id` filter. |
| B17 | `notified` is stored per-vhost, not per-rule | `internal/database/notificationrules.go:117` | `UpdateNotificationRule` sets a top-level `notified` field, so the last rule evaluated overwrites the flag for every other rule on the vhost. | **Medium** | Move to `rules.$.notified`. |
| B21 | `isPresent` never rejects an int, and panics on other types | `internal/config/config.go:222-235` | `case int: return true` means `BASE_PORT=0` etc. always pass validation. The `default` branch `panic`s. | **Medium** | Validate ranges explicitly; return `false` instead of panicking. |
| B22 | `.env` load errors are silently discarded | `internal/config/config.go:93,141-150` | `errors.Is(viper.ConfigFileNotFoundError{}, err)` has its arguments reversed *and* `SetConfigFile` returns `*fs.PathError` for a missing file anyway. The caller then discards the result with `_ =`. A malformed `.env` is invisible. | **Medium** | `errors.As` on the right operand and surface non-"not found" errors. |
| B23 | `ADMIN_GROUPS` is not validated as required | `internal/config/config.go` — `validateConfiguration` | Defaults to empty, which makes `IsAGroupInClaim` fail for everyone — the whole API returns 403 with no obvious cause. Fails closed (good) but is undiagnosable. | **Medium** | Add to `validateConfiguration` or log a loud warning at startup. |

### Security

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| S3 | Raw upstream response body becomes the error | `internal/clients/rest/rest.go:172` | `errors.New(string(bodyBytes))` turns the entire RabbitMQ error page into a Go error that is then surfaced via S2. | **Medium** | Wrap in a typed error with a truncated, sanitised message. |
| S4 | Webhook URLs are user-supplied and unvalidated → SSRF | `internal/helpers/notificationhelper/webhook.go:26-40`, recipient creation in `api/v1/notificationrecipient.go` | The server POSTs to arbitrary URLs supplied through the API, including internal addresses and cloud metadata endpoints. The `/rules/{rule}/test` endpoint makes this trivially on-demand. | **High** | Validate scheme/host, enforce an allowlist or deny RFC1918/link-local, and rate-limit the test endpoint. |
| S5 | Authorization is all-or-nothing; ACL layer is an empty stub | `internal/api/v1/acl.go`, `internal/database/acl.go` and `internal/models/acl.go` are still 1-line package declarations; every handler only calls `IsAGroupInClaim(ctx, rc.AdminGroups)` | There is no per-vhost authorization. Any member of any admin group can read and mutate **every** vhost, rule, recipient and maintenance window. | **High** | Implement the ACL collection (already provisioned in `initCollections` and `mongo-init.js`) and enforce per-vhost scope. |
| S9 | Swagger UI is unauthenticated | `internal/api/routes/v1/apiservice.go` — registered by `SetupUtilityRoutes`, outside the `dex.Authorization()` group | `/api/swagger/*` documents the full API surface to anonymous callers. Note this is now **partly intentional**: the UI must load before the user can complete the OAuth2 flow. | **Low** | Keep the UI public but disable it in production builds, or gate it at the ingress. |
| S10 | No CORS, rate limiting, or security headers | `internal/api/routes/routes.go:66-69` | Only `RequestID`, `Logger`, `Recoverer` and `Timeout` are registered. | **Medium** | Add `cors`, `httprate`, and standard security headers. |
| S11 ⚠️ Partial | RabbitMQ basic-auth credentials may go over plaintext HTTP | `internal/config/config.go` default `http://localhost`; no TLS enforcement in `rest.go` | Credentials are sent in a `Basic` header; with an `http://` host they're in cleartext. | **Medium** | Require `https://` outside local dev; make TLS config explicit. |
| S12 | `gosec` suppressed on the outbound request | `internal/clients/rest/rest.go:158` — `//nolint:gosec // if this causes an exploitation there are bigger issues` | Silences the variable-URL (SSRF) warning that S4 then demonstrates is real. | **Low** | Remove the suppression once URLs are validated. |
| S13 | `WithError(err)` puts raw internal errors in the **client-facing** `error` field | `httpsuite.WriteJSONError` callers throughout `internal/api/v1/` | `WithError` is serialised into the response body, unlike `WithInternalErrorMessage`. `profile.go:40` returns the caller's full group list on a 403, and `dex.go:194` returns the raw OAuth2 exchange error. | **Medium** | Reserve `WithError` for logging only, or scrub before serialising. |

### Resource leaks & HTTP layer

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| R3 | `CheckURLs` opens TCP connections and never closes them | `internal/config/config.go:107-133` | The `net.Conn` returns are still discarded with `_`. Now two connections instead of three (the Prometheus dial is gone), but they stay open for the process lifetime. | **Low** | Assign and `Close()` each. |
| R4 | REST client pins the process-lifetime context onto every request | `internal/clients/rest/rest.go:127,139` | `http.NewRequestWithContext(r.Context, ...)` uses the startup context, so per-request cancellation and deadlines never propagate to RabbitMQ calls. | **Medium** | Pass the caller's `ctx` through each method. |
| R5 ⚠️ Partial (race fixed) | In-memory queue history grows unboundedly | `internal/clients/rabbitmq/rabbitmq.go:22-40` | Now mutex-guarded (good), but the `map[string][]int` keyed by `vhost/queue` is still a package global that is never evicted. Deleted queues leak forever. `historySize` is a hardcoded const. Already flagged by a `TODO` in the file. | **Medium** | Persist to Mongo or add TTL eviction; move `historySize` to config. |
| R7 | `writeJSONResponse` writes a nil body when marshalling fails | `internal/api/httpsuite/response.go:62-75` | Logs the error but continues and writes `nil`, sending a success status with an empty body. | **Medium** | Return 500 and stop. |
| R11 ⚠️ Partial (ordering fixed) | `ReadHeaderTimeout` and `IdleTimeout` are unset | `cmd/unimq/main.go:112-116` | The `WriteTimeout`/handler-timeout inversion is **fixed** (`WriteTimeout` is now 90s vs a 60s `middleware.Timeout`, so the 504 can actually fire). What remains: no `ReadHeaderTimeout` and no `IdleTimeout`, leaving Slowloris exposure. | **Low** | Set both explicitly. |

### MongoDB data layer

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| D2 | Not-found surfaces as **500** on several endpoints | `maintenance.go:99` (expects `ErrMaintenanceNotFound`, gets raw `mongo.ErrNoDocuments`), `notificationrecipient.go:62,212`, `notification.go:150`, `notificationrule.go:383` | Clients cannot distinguish "missing" from "broken". | **Medium** | Map `mongo.ErrNoDocuments` to a sentinel in the DB layer and `errors.Is` it in handlers. |
| D3 | No indexes on any collection | `internal/database/database.go:145-155` | `alarms`, `maintenance_edit_logs.maintenance_id` and all filters are collection scans. | **Medium** | Create indexes at startup. |
| D4 | No pagination or limits on list endpoints | `GetAlarmsAll`, `GetMaintenanceAll`, `GetNotificationsAll`, `GetMaintenanceEditLogs` | Unbounded result sets grow forever (alarms especially). Already noted in `todos.md`. | **Medium** | Add `limit`/`skip`/date filters. |
| D5 | `UpdateNotification` `$set`s the whole struct including `_id` | `internal/database/notifications.go:74-79` | MongoDB rejects mutating the immutable `_id`. Function appears unused — likely broken if ever called. | **Low** | Exclude `_id` from the update document, or delete the function. |

### Dead code / unused features

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| X2 | `CheckURLs` makes every dependency a hard startup gate | `internal/config/config.go:107-133` | A momentary RabbitMQ/Mongo blip at boot kills the process. In Kubernetes this becomes a `CrashLoopBackOff` instead of an unready pod — especially wasteful now that `/readyz` exists. | **Medium** | Warn and retry with backoff; rely on the readiness endpoint instead. |
| X5 | `rest.WithUsername` / `WithPassword` are no-ops | `internal/clients/rest/rest.go:78-88` | Defined but never called; the `Config` fields are never read and auth comes solely from the auth provider. | **Low** | Remove the options. |
| X6 | `AlarmRule.IsTriggered` duplicates the evaluator comparison and is unused | `internal/models/notification.go` | Two sources of truth for the threshold rule. | **Low** | Delete or make `evaluateVhostMetrics`/`evaluateQueueMetrics` call it. |
| X7 | Unused sentinel errors and dead branches | `database/notificationrules.go:31-35` and `api/v1/internal.go:36-41` (identical if/else branches) | Suggests incomplete error handling that was never finished. (`ErrGroupsNotFound` is now used, so that part is closed.) | **Low** | Wire them up or remove. |
| X8 | `AlarmStatus` has six values; only two are ever used | `internal/models/notification.go` | Only `ok` and `firing` are written; `active` is the initial value, `fired`/`inactive`/`unknown` are dead. The state machine is unclear. | **Medium** | Reduce to the states actually used and document transitions. |

### API design & consistency

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| A1 | Response models lack `json` tags → PascalCase keys | `models/notification.go` (`VhostNotification`), `models/alarms.go` (`AlarmEntry`), `models/maintenance.go` (`MaintenanceAdminResponse`, `MaintenanceResponse`), and now `models/health.go:10-14` (`HealthStatus`) | These are returned directly from handlers, so the API emits `Name`/`Recipients`/`Scheduled`/`History`/`AlarmID`, and `/readyz` emits `Database`/`Dex`/`RabbitMQ`, while the rest of the API is `snake_case`. Forces the frontend to match two conventions. | **Medium** | Add `json` tags throughout. |
| A2 ⚠️ Partial | No struct validation despite a validator being available | `models/notification.go`; `httpsuite.IsRequestValid` is called from **one** handler (`notificationrecipient.go:131`) | `PostAlarmRule` accepts negative thresholds and an empty queue name for queue-scoped rules. `validate:` tags exist only in `models/notification.go`. | **Medium** | Add `validate:"required,url,email"` tags and call `IsRequestValid` in every write handler. |
| A4 | Double URL-unescaping of path parameters | 19 call sites across 8 files, e.g. `api/v1/rmq/vhosts.go:77`, `rmq/queues.go`, `api/v1/notificationrule.go` | `net/http` already decodes `r.URL.Path`, so `chi.URLParam` returns a decoded value. `url.QueryUnescape` decodes a second time and also converts `+` to a space — mangling vhost/queue names containing `+` or `%`. Applied inconsistently: `GetRMQVhostUsageHandler` skips it entirely. | **Medium** | Drop the extra unescape; use `chi.URLParam` directly. |
| A5 | ~40 lines of identical boilerplate repeated in 13+ handlers | all of `internal/api/v1/` including the new `rmq/` and `profile/` packages | The group check / param extraction / unescape preamble is copy-pasted, which is exactly how the inconsistencies in D2, A4 and A6 crept in. | **Medium** | Extract an authorization middleware and a param-decoding helper. |
| A6 | Inconsistent status codes for the same class of failure | `maintenance.go` (bad time format → **500**, should be 400), `rmq/vhosts.go:88` and `notificationrecipient.go` (decode/upstream failure → 500 vs 400 elsewhere) | Client error handling can't be uniform. Upstream RabbitMQ failures are reported as 500 although the annotations advertise 502. | **Medium** | Standardise: 400 for input, 404 for missing, 502 for upstream, 500 for internal. |
| A7 | `WithError(err)` passed where `err` is provably `nil` | `notificationrule.go:538,549` (see **G26**), `notificationrecipient.go` | Copy-paste; produces a misleading empty `error` field — confirmed in a live response: `{"error":"","status_code":400,...}`. | **Low** | Remove those options. |
| A8 | Test-notification endpoint rejects email-only vhosts | `api/v1/notificationrule.go:527-534` | Returns **400** when there are no webhook URLs, even if email recipients exist, so email config can't be tested. | **Medium** | Fail only when *no* recipients of any kind exist. |
| A9 | Rule updates use `POST` instead of `PUT`/`PATCH` | `internal/api/routes/v1/apiservice.go:76` | `POST /rules/{rule-id}` for an update is non-RESTful and collides conceptually with rule creation at `POST /rules`. | **Low** | Use `PUT`/`PATCH`. |
| A10 ⚠️ Partial | Swagger annotations don't match the routes | **Fixed** for the `apiservice` routes and for RMQ — routes, handlers and annotations now all agree on `{vhost-name}`/`{queue-id}`, and `swag init` produces zero diff. **Still wrong** for: `status.go:19` documents `/v1/checker/status` but the route is `/api/v1/status`; `health.go` documents `/healthz` and `/readyz` but they are served at `/api/healthz` and `/api/readyz`. | Generated docs and clients are wrong; "Try it out" 404s. | **Medium** | Align the remaining annotations with `internal/api/routes/` and regenerate. |
| A12 | Duplicate recipients are accepted | `api/v1/notificationrecipient.go:138-146` | No uniqueness check, so the same webhook can be added N times and receives N copies of every alarm. Listed in `todos.md`. | **Low** | Deduplicate on URL/email before `$push`. |

### Notification checker specifics

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| N1 | State diverges when the status update fails | `internal/notify/checker.go:240-246` | If `UpdateNotificationRule` errors, the code only logs and continues to log the alarm and notify. The DB still says `ok`, so the alarm re-fires and re-notifies every interval. | **Medium** | Return early on update failure. |
| N2 | Only `>=` comparisons are supported | `internal/notify/evaluators.go:73,104` | Rules cannot express "below threshold" (e.g. throughput stalled). `no_consumer` is special-cased instead, which is why it needed its own early return. | **Medium** | Add a comparison operator to `AlarmRule`. |
| N3 | Checker interval is hardcoded | `cmd/unimq/main.go:98` | `WithInterval(60*time.Second)` is not configurable via env, unlike every other tunable. | **Low** | Add `CHECKER_INTERVAL_S`. |
| N4 | `GetMetrics` fetches **all** cluster connections and channels per vhost | `internal/clients/rabbitmq/rabbitmq.go:226-260` | Two full cluster-wide listings per vhost per tick, filtered client-side. With many vhosts this is O(vhosts × cluster size) every 60s and will hammer the management API. | **Medium** | Use `/vhosts/{vhost}/connections` / `/channels`, or fetch once per tick and group. |
| N5 | A rule type is signalled as an error | `internal/notify/evaluators.go:21-23` | `AlarmTypeMaintenance` returns `ErrNotificationRuleInMaintenance` from `EvaluateMetrics`, conflating "not applicable" with failure. The same applies to `ErrNotificationRuleDisabled`. | **Low** | Skip these explicitly rather than via an error. |
| N6 | One-shot startup delay uses a `Ticker` | `internal/notify/checker.go:87` | A `Ticker` is created for a single 15s wait. | **Low** | Use `time.NewTimer`/`time.After`. |
| N7 | `sendEmail` ignores port, username and password | `internal/helpers/notificationhelper/email.go:56` | `mail.NewClient(config.EmailSMTPHost)` only — `EMAIL_SMTP_PORT/USERNAME/PASSWORD` are dropped, so authenticated or non-default-port SMTP fails (it dials `:25`). A new client is also constructed per message. The guard checks `EmailFromAddress` but reports "SMTP server is not configured". | **Medium** | Pass the full config via `mail.WithPort`/`WithUsername`/`WithPassword` and reuse one client. |
| N8 | Webhook payload is Slack-shaped and hardcoded | `internal/helpers/notificationhelper/webhook.go:29-31` | `{"text": ...}` works for Slack/Teams but not generic webhooks, and the `json.Marshal` error is discarded with `_`. | **Low** | Make the payload template configurable; handle the error. |
| N9 | `BuildMessage` switches on string literals, not the `AlarmType` constants | `models/notification.go` | Renaming a constant silently falls through to the generic message. | **Low** | Switch on the typed constants. |
| N10 | `SendWebhook` ignores the caller's context | `internal/helpers/notificationhelper/webhook.go:31` | It derives its 10s timeout from `context.Background()` rather than the passed `ctx`, which is used only for logging. `SendWebhooks` takes no `ctx` at all. Checker shutdown therefore can't cancel in-flight webhooks. `SendEmails` accepts a `ctx` and never uses it. | **Low** | Derive from the caller's context and thread it through `SendWebhooks`. |
| N11 | `EmailSenderInstance` is a package-level global | `internal/helpers/notificationhelper/email.go:73` | Set once in `main.go:93`. Any code path that runs without `InitEmailSender` — notably tests — nil-derefs inside `SendEmails`. This is the main obstacle to testing the notification path. | **Medium** | Inject the sender into `Checker` and `APIService` instead. |

### Observability & testing

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| O1 | Test coverage is minimal | tests exist only in `internal/database`, `internal/notify`, `internal/api/httpsuite`, `internal/helpers/notificationhelper` | No tests for `api/v1` (the bulk of the logic), `config`, `clients/*`, or `logger`. Most bugs above are trivially testable — **G18–G21 would all have been caught by a single handler/route test**. Already in `todos.md`. | **Medium** | Add handler tests with a mock RMQ/DB, and a route-tree test that asserts every registered path resolves. |
| O2 | Logs are `TextHandler`, not JSON | `internal/logger/logger.go:29` | Harder to parse in Kubernetes log aggregation. | **Low** | Use `slog.NewJSONHandler` (optionally env-switched). |
| O4 | Unchecked assertion in the log handler | `internal/logger/logger.go:56` | `a.Value.Any().(*slog.Source)` will panic if the attr type ever differs. | **Low** | Use the comma-ok form. |
| O5 | `slog.Error` used for an informational message | `internal/logger/logger.go:38` | "updating log level" is logged at Error. | **Low** | Use `slog.Info`. |

---

## AUTHENTICATION

Swagger now drives an OAuth2 **authorization code + PKCE** flow against Dex (`unimq-swagger` public client, `trustedPeers` on `unimq-dashboard` so one token is accepted by both). The resource-owner password flow and `LoginHandler` were removed. The items below are what remains.

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| U1 | `state` is a hardcoded literal and is never validated | `internal/clients/dex/dex.go:163` (`AuthCodeURL("state", ...)`) and `OauthCallbackHandler`, which never reads `state` | This is the CSRF defence for the authorization-code flow. A fixed, unchecked value provides none. Becomes critical once login moves from the frontend to the backend. | **High** | Generate a random `state` per request, store it in a short-lived `HttpOnly` cookie, and compare it in the callback. |
| U2 | The `id_token` is returned in a redirect **query string** | `internal/clients/dex/dex.go:212-213` — `http.Redirect(w, r, "/?token="+idToken, 302)` | Query strings land in browser history, `Referer` headers, proxy and web-server access logs. A leaked `id_token` is a full session. | **High** | Set an `HttpOnly; Secure; SameSite=Lax` cookie, or return it via a fragment the SPA consumes and strips. |
| U3 | The frontend still holds an OIDC **client secret** | `web-src/src/auth/auth.config.ts` (see **F1**) | Unchanged this pass. The backend-side flow exists (`/login/redirect`, `/login/callback`) but the frontend has not been migrated onto it yet. | **High** | Complete the migration to the backend flow, then make the frontend client public. |
| U4 | Production Dex needs matching registration | `volumes-for-compose/dex-config-dev.yaml` only covers local dev | The `unimq-swagger` public client, its redirect URIs, the `trustedPeers` entry on `unimq-dashboard`, and `web.allowedHeaders` (`Content-Type`) must all exist on the shared Dex too. Without `allowedHeaders`, the CORS **preflight** for the token request is rejected and Swagger UI reports only an opaque `TypeError: NetworkError`. | **Medium** | Register the client and headers in the production Dex config; document the requirement in the README. |
| U5 | The Swagger issuer is rewritten via a hardcoded sentinel | `internal/api/routes/routes.go:27-37` | `defaultSwaggerOIDCIssuer` must stay byte-identical to the `@securitydefinitions.oauth2.accessCode` URLs in `cmd/unimq/main.go`. If either is edited alone, the rewrite silently no-ops and Swagger points at `localhost:5556` in production. | **Medium** | Derive both from one constant, or post-process the spec by JSON path instead of string replacement. |
| U6 | `GetProfileHandler` requires admin group membership | `internal/api/v1/profile/profile.go:37-41` | A user must be in `ADMIN_GROUPS` just to read **their own** profile, so the endpoint can't serve non-admin users once S5 introduces scoped access. It also returns the error via `WithError`, leaking the caller's full group list (see **S13**). | **Medium** | Require only a valid token; drop the group check. |

---

## FRONTEND

Stack: React 19 + TypeScript, Vite (multi-entry, **no client-side router**), Tailwind v4 + shadcn/Radix, `oidc-client-ts` via `react-oidc-context`, native `fetch` (no React Query/SWR).

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| F1 | OIDC **client secret** is shipped to the browser | `web-src/.env.local.example:5-7`, `web-src/public/env.js.tmpl:1-5`, `web-src/src/auth/auth.config.ts:10-13` | A secret delivered to the browser is public. Matches the open item at the top of `todos.md`. | **High** | Use a public client with PKCE, or move the code exchange to the backend. |
| F2 | Mutations never check `res.ok` | `web-src/src/services/notifications.ts:35-95`, `web-src/src/services/maintenance.ts:34-51` | `addRule`, `toggleRule`, `deleteRule`, `addRecipient`, `addMaintenance` resolve successfully on 400/403/500. The UI closes the form or reloads as if the change succeeded. | **High** | Add a shared `throwIfNotOk` helper used by every mutation. |
| F3 | `getVhosts` parses error responses as data | `web-src/src/services/vhosts.ts:10-14` | No `res.ok` check before `res.json()`; an error response yields a misleading empty vhost list. | **High** | Check `res.ok` and throw a parsed error. |
| F4 | No central 401/403 handling | `web-src/src/lib/apiClient.ts:7-14` | `apiFetch` attaches the token but never handles expiry — no silent renew, no redirect to login. Sessions expire into blank pages. | **High** | Intercept 401, attempt renewal, else redirect through the OIDC provider. |
| F5 | Unhandled promise rejections in data hooks | `web-src/src/hooks/useVhostNotification.ts:24-55`, `useMaintenance.ts:11-49`, `useClusters.ts:9-13` | `.finally()` without `.catch()`. API failures produce unhandled rejections and a UI that looks empty rather than errored. | **High** | Add `catch` and expose an explicit error state. |
| F6 | Optimistic alarm toggle is never reconciled | `web-src/src/components/notifications/EditAlarm.tsx:47-57` | `toggleRule` isn't awaited or caught; a failed toggle leaves the UI showing the wrong state. | **High** | Await, revert on failure, surface the error. |
| F7 | The `/queue` page is a blank `<div />` | `web-src/src/pages/Queue.tsx:6-8`, registered as a build entry in `vite.config.ts:16` | A publicly reachable, authenticated, entirely non-functional page ships in production. | **High** | Implement it or remove the entry. |
| F8 | `package.json` declares `"scripts"` **twice** | `web-src/package.json:6` and `:44` | JSON parsers keep only the second object, silently discarding the earlier `test` and `previte` scripts — so `npm test` is effectively unavailable. | **Medium** | Merge into one `scripts` object. |
| F9 | Stale-response races in hooks | `useVhostNotification.ts:21-55`, `useMaintenance.ts:9-49`, `useClusters.ts:8-13` | No `AbortController` or cancellation flag (unlike `useQueues`). Under StrictMode or fast navigation, stale results overwrite current state. | **Medium** | Abort on cleanup or guard with a mounted flag. |
| F10 | Dashboard maintenance edit link drops the vhost | `web-src/src/components/dashboard/DashboardMaintenanceWidget.tsx:108` | Builds `/maintenance/edit?id=...` without the vhost that other navigation preserves, so editing can land in the wrong context. | **Medium** | Include the encoded vhost. |
| F11 | `NotifyRule` loads one vhost and navigates back to another | `web-src/src/pages/NotifyRule.tsx:10-20` | Loads with `params.get('vhost')` falling back to `selected`, but the back link always uses `selected`. | **Medium** | Derive a single validated vhost value. |
| F12 | Missing env config degrades to empty strings | `web-src/src/env.ts:9-12`, `auth.config.ts:7-13` | An absent issuer or client ID produces confusing runtime OIDC failures instead of a clear error. | **Medium** | Validate at startup and render a configuration error. |
| F13 | Unchecked casts of `FormData` and OIDC claims | `components/notifications/AlarmCard.tsx:187-191`, `RecipientCard.tsx:83-87`, `components/profile/Profile.tsx:14-27` | `FormData.get()` can return `null` or `File`; claims are untrusted external data. Casts provide no runtime safety. | **Medium** | Validate with guards or Zod. |
| F14 | Response typings don't match the backend | `web-src/src/services/maintenance.ts:13-31` | `ApiResponse<Maintenance, Maintenance>` is reused for a list body that actually contains `Scheduled`/`History` (PascalCase — see A1). Unchecked assertions hide the drift. | **Medium** | Define per-endpoint types once A1 is fixed. |
| F15 | No error/loading distinction on main pages | `pages/index.tsx:50-56`, `pages/Maintenance.tsx:10-29`, `pages/Notifications.tsx:11-31` | Users can't tell "no records" from "API is down" or "forbidden". | **Medium** | Render error states from every hook. |
| F16 | Deletes use full-page reloads and untracked timers | `components/notifications/DeleteItem.tsx:31-43`, `components/maintenance/DeleteMaintenance.tsx:27` | `setTimeout` isn't cleaned up on unmount; the UI reloads the whole page instead of updating state. | **Low** | Update parent state and clear timers on unmount. |
| F17 | `localStorage` is trusted as the declared type | `web-src/src/hooks/useLocalStorage.ts:5-9` | `JSON.parse(item) as T` accepts corrupt or outdated values, making schema changes unsafe. | **Low** | Validate and migrate/reset. |
| F18 | Dashboard preferences are shared across users | `web-src/src/hooks/useDashboard.ts:20-23` | The `dashboard-widgets-v1` key isn't namespaced, so a second user on the same browser profile inherits the first user's layout. | **Low** | Namespace by authenticated `sub`. |
| F19 | Multi-entry build instead of a router | `web-src/vite.config.ts:13-30` | Every page re-bootstraps its own `AuthProvider`; navigation, 404s and caching are all bespoke. A `TODO` at line 30 flags a dev-only proxy `bypass` workaround that doesn't match production serving. | **Low** | Consider a single entry with a router, or document and smoke-test the multi-entry setup. |

---

## HELM

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| H1 | Plaintext credential defaults in version control | `unimq/values.yaml:59-88` | Default MongoDB/RabbitMQ/OIDC values ship in the chart. Installing without overrides creates predictable credentials. | **High** | Require an external Secret; remove credential defaults. |
| H2 | Chart generates a Secret from insecure placeholders | `unimq/templates/secret.yaml:9-13` | Materialises `"password"`, `"guest"`, `"placeholder"` into a real Secret (base64 is not encryption). | **High** | Make `existingSecret` mandatory in production and fail rendering on placeholder values. |
| H4 | Probes are TCP-only | `unimq/templates/backend-deployment.yaml:52-61`, `frontend-deployment.yaml:61-70` | The backend can accept TCP while RabbitMQ/Mongo are unreachable, so a broken pod stays in service. `/healthz` and `/readyz` **now exist** (at `/api/healthz` and `/api/readyz`), so there is no longer any reason for TCP probes — but note `/readyz` currently returns malformed JSON (**G19**). | **Medium** | Switch to `httpGet` against `/api/readyz` once G19 is fixed. |
| H5 | One `replicaCount` drives both deployments, defaulting to 1 | `unimq/values.yaml:4`, both deployment templates line 10 | Backend and frontend have different scaling needs; a single backend replica is a SPOF. Combined with X2, a dependency blip takes the whole service down. | **Medium** | Split the values; add optional HPA/PDB. |
| H6 | Resource requests/limits are hardcoded | `backend-deployment.yaml:40-44`, `frontend-deployment.yaml:45-49` | Cannot be tuned per environment; risks throttling or OOM kills. | **Low** | Move into `values.yaml`. |
| H7 | Namespace is hardcoded and not created by the chart | `unimq/values.yaml:3` and all templates | Install fails unless the namespace is pre-created. | **Low** | Document the prerequisite or template it optionally. |
| H8 | No `_helpers.tpl`, no standard labels/selectors | `unimq/templates/` | Missing the conventional `app.kubernetes.io/*` labels and name/fullname helpers makes selectors brittle. | **Low** | Add a helpers template and standard label blocks. |

---

## DOCKER / CONTAINERS

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| C1 | Compose ships default credentials | `docker-compose.yaml:24-25,49,68-71` | RabbitMQ `guest/guest`, MongoDB `admin/adminpassword`, Mongo Express `test/test`. | **High** (with C2) | Source from an ignored `.env`; fail if absent. |
| C2 | All dev services are published on every host interface | `docker-compose.yaml` — `dex`, `rabbitmq`, `mongodb`, `mongo-express` | These are reachable from the local network, with the credentials in C1. (Prometheus is no longer among them — it was removed from Compose.) | **High** | Bind to `127.0.0.1`. |
| C3 | Frontend build output path is confusing and fragile | `web-src/vite.config.ts:13` (`outDir: "../web-src/static/dist"`) and `dockerfiles/Dockerfile.frontend:12` (`COPY --from=builder /web-src/static/dist`) | The `../web-src/` prefix means the output lands in `web-src/static/dist` locally but at the container **root** `/web-src/static/dist` in the builder (where `WORKDIR` is `/app`). It works only by coincidence of the two mistakes cancelling out. | **Medium** | Set `outDir: "static/dist"` and copy from `/app/static/dist`. |
| C4 | Mongo Express connection URL has a stray trailing quote | `docker-compose.yaml:70` | `ME_CONFIG_MONGODB_URL` ends with `unimq"`, likely breaking the connection. | **Medium** | Remove the quote. |
| C5 | Base images are not pinned | `Dockerfile.frontend:3` (`node:22-alpine`), `Dockerfile.backend:17` (`gcr.io/distroless/static`, no tag) | Rebuilds are not reproducible and can silently pull new vulnerabilities. | **Medium** | Pin to digests or exact patch versions. |
| C6 | Unused `GCR_MIRROR` build arg | `Dockerfile.frontend:1` | Declared but never referenced. | **Low** | Use it or remove it. |
| C7 | Mongo Express doesn't wait for MongoDB health | `docker-compose.yaml:73-74` | Uses plain `depends_on` despite MongoDB having a healthcheck, so it can start too early and fail. | **Low** | Use `condition: service_healthy`. |
| C8 | Only MongoDB has a healthcheck | `docker-compose.yaml` | Startup ordering and local debugging are unreliable for Dex and RabbitMQ. | **Low** | Add healthchecks. |
| C9 | No `HEALTHCHECK` in either image | `dockerfiles/Dockerfile.backend`, `Dockerfile.frontend` | No health signal outside Kubernetes. | **Low** | Add one, or document that orchestrator probes are required. |

---

## CI/CD

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| P1 | Releases publish a mutable `latest` tag | `.github/workflows/release_backend.yaml:23`, `release_frontend.yaml:23` | Deployments pinned to `latest` change without a manifest change and aren't reproducible. | **Medium** | Deploy immutable version/SHA tags; treat `latest` as an alias only. |
| P2 | `swag` installed from `@latest` in CI | `.github/workflows/testandbuild_backend.yaml:37` | Unpinned tooling can break builds or introduce supply-chain risk. | **Medium** | Pin the version. |
| P3 | Thin CI verification | `testandbuild_backend.yaml:33`, `testandbuild_frontend.yaml:27` | Backend runs `go test ./...` against very few tests (O1); the frontend appears to only build — no tests, no dependency audit, no security scan. | **Low** | Add frontend tests, `govulncheck`, and dependency auditing. |

---

## DEVELOPMENT_ENVIRONMENT

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| E1 | `.env.example` documents six variables the backend doesn't read | `.env.example:25-30` — `RABBITMQ_CHANNEL_LIMIT`, `RABBITMQ_CONNECTION_LIMIT`, `RABBITMQ_CONNECTION_MAX`, `RABBITMQ_QUEUE_LIMIT`, `RABBITMQ_QUEUE_MAX`, `RABBITMQ_MESSAGE_UNACK_LIMIT` | **None** of these are bound in `internal/config/config.go` any more — only `RABBITMQ_HOST/PORT/USERNAME/PASSWORD` are. Static limits were removed in favour of reading them from RabbitMQ, but the example was never cleaned up. Developers will set values that do nothing. | **Medium** | Delete the six dead entries. |
| E3 | `MONGODB_HOST` semantics are ambiguous | `.env.example:9` (`"localhost"`) vs `config.go:59` default (`"mongodb://localhost"`) | The two disagree on whether a scheme is expected; `CheckURLs` strips it, `CreateUri` escapes it. | **Medium** | Document and enforce one format. |
| E4 | Local `.env` uses predictable credentials | `.env` (gitignored) | Low risk on its own, but dangerous combined with C2 (services bound to all interfaces). Confirmed **not committed**. | **Low** | Generate local-only credentials. |
| E5 | `.env.example` ships a literal OIDC client secret | `.env.example:48` — `OIDC_CLIENT_SECRET = "O8jad9kl1PD1er"` | It's only the dev Dex secret, but a committed example that looks like a real credential invites copy-paste into production and trips secret scanners. | **Low** | Replace with an obvious placeholder such as `"changeme"`. |

### Verified as NOT issues

- **No secrets or database files are committed.** `git ls-files volumes-for-compose` returns only config files. No MongoDB data directory is tracked.
- **`.env` is correctly gitignored** and absent from the index.
- **`go build ./...`, `go vet ./...` and `go test ./...` all pass** as of `fc6b9ea` (run as three separate commands — chaining them with `&&` is what produced a false green report in the fifth pass).
- **`new(float64(x))` is not a compile error.** Go 1.27 permits `new(expr)`; an earlier draft of this document wrongly flagged it.
- **`volumes-for-compose/prometheus-config.yaml` was deleted**, not left stale. Only `dex-config-dev.yaml` and `mongodb/mongo-init.js` are tracked; the local `volumes-for-compose/data/` directory is untracked.

---

## FIXED — CHANGE LOG

Everything below has been verified fixed in the current tree, and has been removed from the backlog tables above. Locations are where the fix lives **today**, after the seventh-pass restructuring (`httpsuite` moved to `internal/api/httpsuite`, routes to `internal/api/routes`, RabbitMQ handlers to `internal/api/v1/rmq`). Rows marked *(partial)* are still open in the backlog.

### Backend — correctness

| # | What was wrong | How it was fixed | Where it lives now |
|---|----------------|------------------|--------------------|
| **B1** | Alarm history kept only the **first** event per rule. `AddAlarm` did an `InsertOne` keyed on the rule ID (`AlarmID` is `bson:"_id"`), so every later fired/resolved transition failed with a duplicate-key error that was only logged. | Switched to `$push` into the `entries` array with `SetUpsert(true)`, so the first event creates the document and later ones append. `AlarmEntry` is only `_id` + `entries`, so the upserted document is complete. | `InsertAlarmEntries` — `internal/database/alarms.go:70`; called from `checkRule` — `internal/notify/checker.go:298` |
| **B2** | `checkMaintenanceSchedules` ran *inside* the per-vhost loop but marked entries `Notified=true` globally, so only the first vhost's recipients ever got a maintenance notice. | Moved the call out of the loop; `urls`/`emails` are accumulated across all vhosts first, then the maintenance check runs once per tick. | `runChecks` — `internal/notify/checker.go:150-180` |
| **B4** | `GetMaintenanceHandler` called `WriteJSONError` on history failure and then **fell through** to `SendResponse`, writing two responses (`superfluous WriteHeader`, corrupt double-JSON body). | Added the missing `return` after the error response. | `GetMaintenanceHandler` — `internal/api/v1/maintenance.go` |
| **B6** | `NewRMQClient` returned `&RMQClient{restClient: ...}` and never assigned `Limits`, so `WithRMQLimits` and all `RABBITMQ_*_LIMIT` settings were dead. | `Limits` is now assigned from the supplied options. | `NewRMQClient` — `internal/clients/rabbitmq/rabbitmq.go` |
| **B7** | `CreateUri` ran `url.QueryEscape` on the whole host. The default `MONGODB_HOST` is `"mongodb://localhost"`, which escaped to `mongodb%3A%2F%2Flocalhost`, producing an unusable URI. | Strips a leading `mongodb://` with `strings.TrimPrefix` before building the URI. | `CreateUri` — `internal/database/database.go:150-168` |
| **B9** | `AlarmRuleUpdate` was decoded into a zero-valued struct and threshold *and* message were always written, so omitting `message` blanked it and omitting `threshold` set it to `0`. | Fields changed to pointers so "absent" and "explicitly zero" are distinguishable; only supplied fields are written. | `AlarmRuleUpdate` — `internal/models/`, applied in `internal/api/v1/notificationrule.go` |
| **B10** | In `ensureNotificationHostExists` the `errors.Is(err, mongo.ErrNoDocuments)` block had no `else` and the outer `if err != nil` never returned, so genuine DB failures were swallowed and fell through to a second query. | Added an `else` branch returning the wrapped generic error, leaving the `ErrNoDocuments` path free to continue to creation. | `ensureNotificationHostExists` — `internal/api/v1/internal.go:47-53` |
| **B11** | `UpdateLogLevel` returned early when `slog.Default().Enabled(ctx, level)` was true. Since the default is Info, Warn and Error were already "enabled", so `LOG_LEVEL` could never be made *less* verbose. | Removed the guard; the handler is now always rebuilt. | `UpdateLogLevel` — `internal/logger/logger.go:36` |
| **B12** | Logs never carried `request_id` — the handler read `ctx.Value("request_id")` with a raw string key, but chi stores it under its own unexported key type. | Reads `middleware.GetReqID(ctx)`. | `internal/logger/logger.go` |
| **B13** | `middleware.RequestID` was registered *after* `middleware.Logger`, so IDs were absent from log output even once B12 was fixed. | Reordered the middleware chain so `RequestID` comes first. | `SetupRoutes` — `internal/routes/routes.go:50-51` |
| **B18** | `UpdateNotificationRule` wrote `rules.$.lastResolved`, a field that does not exist on the `AlarmRule` model — schema drift, written but never read. | The write was dropped. No `lastResolved`/`LastResolved` references remain anywhere in the tree. | `UpdateNotificationRule` — `internal/database/notificationrules.go:120` |

### Backend — security

| # | What was wrong | How it was fixed | Where it lives now |
|---|----------------|------------------|--------------------|
| **S1** | `aClaimGroups.([]any)` was an unchecked type assertion. Any authenticated token carrying `groups` as a string (or anything non-array) panicked — remotely triggerable on every endpoint. | Uses the comma-ok form and returns a new `ErrInvalidGroupsType` sentinel instead of panicking. | `internal/routes/httpsuite/context.go` |

### Backend — resource handling / lifecycle

| # | What was wrong | How it was fixed | Where it lives now |
|---|----------------|------------------|--------------------|
| **R1** | The `StatusCode > 399` branch in `RestClient.request` read the body and returned without closing it, leaking a connection on every RabbitMQ error. | A single `defer resp.Body.Close()` now covers all return paths — and, after a follow-up fix, is correctly registered **after** the `err != nil` check, so a failed `Do` no longer nil-derefs. | `RestClient.request` — `internal/clients/rest/rest.go:152-168` |
| **R8** | The Mongo client was never disconnected on shutdown. | A `Close` method was added and is called in the shutdown path. | `internal/database/database.go`, invoked from `cmd/unimq/main.go` |
| **R9** | `wg.Add(1)` **and** `wg.Go(...)` **and** `defer wg.Done()` were all used together — adding 2 and doning 2, balancing only by accident. | Uses `wg.Go(f)` alone. | `cmd/unimq/main.go:111` |
| **R10** | The server goroutine assigned `err = server.ListenAndServe()` while main later assigned `err = server.Shutdown(...)` — a data race on a shared variable. | `err` is now scoped to the goroutine. | `cmd/unimq/main.go` |
| **R11** *(partial)* | `WriteTimeout` (30s) was shorter than `middleware.Timeout` (60s), so the write deadline always fired first and the handler timeout could never return its 504. | `WriteTimeout` raised to 90s, comfortably above the handler timeout. `ReadHeaderTimeout` and `IdleTimeout` are **still unset** — this row is not fully closed. | `cmd/unimq/main.go:112` |

### Backend — API surface

| # | What was wrong | How it was fixed | Where it lives now |
|---|----------------|------------------|--------------------|
| **A11** | The only status endpoint was `/api/v1/status`, behind auth, so Kubernetes probes had nothing usable to hit. | Added unauthenticated `/api/healthz` and `/api/readyz`. Also unblocks **H4** (chart still uses TCP-only probes). | `internal/routes/` |

### Regressions introduced during fixing, since resolved

| # | What went wrong | How it was resolved | Where it lives now |
|---|-----------------|---------------------|--------------------|
| **G2** | `defer resp.Body.Close()` was registered **before** the `if err != nil` check, so any failed outbound RabbitMQ call nil-dereffed — fatal in the checker goroutine, which has no `Recoverer`. | The error check was moved above the `defer`. | `internal/clients/rest/rest.go:154-168` |
| **G3** | `url.QueryEscape` → `url.PathEscape` on Mongo credentials. `PathEscape` does not escape `@`, so `p@ssw0rd` produced an unparseable URI and the unit test went red. | Userinfo is now built with `url.UserPassword(username, password).String()`, which escapes correctly. `TestDatabaseConnection` passes. | `CreateUri` — `internal/database/database.go:157-168` |
| **G4** | The generic-failure `return` added for B10 sat outside the `ErrNoDocuments` branch, so the path that successfully *created* a notification host fell through into the error return — first-ever rule creation always 500'd. | Moved into an `else` branch. | `internal/api/v1/internal.go:50-52` |
| **G5** | `urls` accumulated across vhosts but was still passed to `checkRule` inside the loop, so vhost N's alarms were also delivered to vhosts 1..N-1's webhooks. | `checkRule` is now passed `vhost.WebhookURLs()` directly; the accumulated slices are used only by `checkMaintenanceSchedules`. | `runChecks` — `internal/notify/checker.go:175` |
| **G6** | `InsertAlarmEntries` used `$push` with no upsert, so the parent document was never created and every write was a silent no-op. | Added `options.UpdateOne().SetUpsert(true)`. | `internal/database/alarms.go:70-74` |
| **G7** | `Checker.EmailConfig` was read but had no setter and was never wired in `main.go`, so it was always nil and `SendEmail` dereferenced it immediately. | Added the `WithEmailConfig` option and wired it in `main.go`; `SendEmail` nil-guards and returns the `ErrEmailNotConfigured` sentinel, which the maintenance loop handles with a warning and `continue`. | `WithEmailConfig` — `internal/notify/checker.go:65`; `cmd/unimq/main.go:95`; `SendEmail` — `internal/helpers/notificationhelper/notificationhelper.go:41-46` |

### Added in the fourth pass

| # | What was wrong | How it was fixed | Where it lives now |
|---|----------------|------------------|--------------------|
| **S6** | `AuthorizeScope` had an inverted comparison: `ScopeAdmin = 0` is the *highest* privilege but the check was `allowed <= scope`, so `AuthorizeScope(ScopeRead, ScopeAdmin)` returned true and would have granted a read-only user admin rights. | The whole scope/ACL block was deleted. It had no callers, so removing it eliminates the latent privilege-escalation bug outright. Note this does **not** address **S5** — per-vhost authorisation is still unimplemented and `acl.go` is now an empty package stub. | `internal/models/acl.go` (now a package declaration only) |
| **R6** | `writeJSONResponse` called itself on write failure, so a broken connection produced unbounded recursion plus a second `WriteHeader`. | The recursive error branch was replaced with a log-and-return. | `writeJSONResponse` — `internal/routes/httpsuite/response.go:71-75` |

### Added in the fifth pass

| # | What was wrong | How it was fixed | Where it lives now |
|---|----------------|------------------|--------------------|
| **B14** | `MaintenanceEntry.UnmarshalJSON` validated `status` but never assigned it, and force-reset `Notified`, so every decoded entry came back with an empty status and `Notified=false`. | `e.Status = ParseMaintenanceStatus(aux.Status)` and `e.Notified = aux.Notified` are now assigned. | `MaintenanceEntry.UnmarshalJSON` — `internal/models/maintenance.go:147-148` |
| **B15** | Nothing rejected a maintenance window whose end was before its start; `AdvanceMaintenanceStatuses` then immediately flipped it to `done`. | An `end.Before(start)` check was added to both the create and patch paths. | `PostMaintenanceEntry.ToMaintenanceEntry` — `internal/models/maintenance.go:81`; `PatchMaintenanceEntry.Validate` — `internal/models/maintenance.go:193` |
| **S2** | `newErrorResponse` copied `err.Error()` into the client-facing `ErrorResponse`, leaking raw Mongo/RabbitMQ driver text, hostnames and internals to any caller. | The branch that set `Error` was removed, so responses carry only the curated message. The underlying error is still logged server-side by `WriteJSONError`/`WriteJSONErrorForbidden`, so nothing is lost for debugging. | `newErrorResponse` — `internal/routes/httpsuite/error.go:67-72` |
| **G1a** | The webhook error branch formatted `req.Response.StatusCode`; `req.Response` is nil on a client request, so reporting a failing webhook panicked the checker. | Reads `resp.StatusCode`, and logs the failing URL and status. | `SendWebhooks` — `internal/helpers/notificationhelper/notificationhelper.go:44-47` |
| **G1b** | The `err` check after `http.NewRequestWithContext` was missing, so a malformed, user-supplied webhook URL dereferenced a nil `req`. | The guard is restored: the error is logged, wrapped into `lastErr`, and the loop continues. | `SendWebhooks` — `internal/helpers/notificationhelper/notificationhelper.go:23-27` |
| **G8** | Adding a recipient always returned 400. `IsRequestValid` was passed the `*http.Request` rather than the payload, and its `*ValidationErrors` return was assigned to an `error` variable, where a nil pointer is not a nil interface. | `IsRequestValid` now returns `error` and returns a literal `nil` on success, and the handler passes the decoded `recipient` struct. Both halves of the typed-nil trap are gone. | `IsRequestValid` — `internal/routes/httpsuite/validation.go:47`; handler — `internal/api/v1/notificationrecipient.go:129` |
| **G9** | `http.MaxBytesReader`'s return value was discarded, so the 10 MB request-body limit had no effect and **S8** remained open. | The returned reader is captured as `closer` and the JSON decoder reads from it, so the limit is enforced. This closes **S8**. | `ReadResponse` — `internal/routes/httpsuite/response.go:88-91` |
| **D2** *(partial)* | Not-found conditions surfaced as 500 because handlers compared against raw `mongo.ErrNoDocuments`. | Sentinel errors were introduced (`ErrAlarmNotFound`, `ErrMaintenanceNotFound`, `ErrRecipientNotFound`, `ErrNotificationRuleNotFound`, `ErrVhostNotFound`) and are wrapped at the data layer, with 6 handler sites mapping them to 404. Not yet applied uniformly — this row stays open. | `internal/database/*.go`; consumers in `internal/api/v1/` |

### Added in the sixth pass

| # | What was wrong | How it was fixed | Where it lives now |
|---|----------------|------------------|--------------------|
| **B3** | Email recipients never received alarm notifications — `NotifyAlarm` was webhook-only, so `type: "email"` recipients were silently ignored for alarms. | `NotifyAlarm` now iterates `vhost.EmailRecipients()` and sends via the email sender, treating `ErrEmailNotConfigured` as a warning rather than a hard error. | `NotifyAlarm` — `internal/notify/checker.go:328-345` |
| **B5** | Maintenance timestamps were parsed inconsistently, so a window's meaning depended on the server's local timezone. | A `timehelper` package centralises parsing: `ParseTimeInUTC` uses `time.ParseInLocation(layout, value, time.UTC)`. All call sites now use it — no bare `time.Parse` remains outside tests. | `internal/helpers/timehelper/timehelper.go`; used across `internal/models/maintenance.go` |
| **B19** | `updated_at` was written as a string while `start`/`end` were BSON dates, so range queries and sorting on it could not work. | `UpdatedAt` is now `time.Time` on both `MaintenanceEntry` and `MaintenanceEditLog`, and the database layer writes a real time value. | `internal/models/maintenance.go:104,208`; `internal/database/maintenance.go:304` |
| **G10** | `fmt.Sprintf` was called with the `%w` verb, breaking `go vet` and emitting `%!w(*errors.errorString=...)` to the client. | Replaced with a correct format verb. `go vet ./...` passes again. | `internal/api/v1/maintenance.go` |
| **G11** | `InsertAlarmEntries` treated `UpsertedCount == 0` as "not found", so every append to an existing alarm reported a spurious failure. | The check is now `results.UpsertedCount == 0 && results.MatchedCount == 0`, which is only true when nothing was created *and* nothing matched. | `InsertAlarmEntries` — `internal/database/alarms.go:115` |
| **G12** | Thirteen sites treated `ModifiedCount == 0` as "not found", so writing a value a field already held returned an error — re-enabling an enabled rule, re-saving an unchanged threshold, and re-marking a maintenance entry all failed. | All of them now test `MatchedCount == 0`, matching the pattern `notifications.go` already used. `ModifiedCount` is retained only for logging. | `internal/database/notificationrules.go`, `notificationrecipients.go`, `maintenance.go` |
| **G13** | The `Recipient` validation tags were spelled `validation:"..."`, which go-playground/validator ignores. | Renamed to `validate:"..."`. | `internal/models/notification.go:45-49` |

### Added in the seventh pass

| # | What was wrong | How it was fixed | Where it lives now |
|---|----------------|------------------|--------------------|
| **B8** | Queue-scoped rules evaluated a missing queue as `0` instead of erroring. `v := new(float64)` seeded a non-nil pointer, so the `if v == nil` guard was dead code and a rule with `threshold <= 0` fired a false alarm for a deleted or renamed queue. Open across three previous attempts. | The evaluator was extracted from `checker.go` into its own file and rewritten around an explicit `queueFound` flag. When no queue matches, it returns `ErrNotificationRuleQueueNotFound` instead of a zero value, and `v` is now declared `var v *float64`. | `evaluateQueueMetrics` — `internal/notify/evaluators.go:78-108` |
| **B20** | The delete error message interpolated the `id` constant instead of the vhost, producing `notification not found for vhost _id.` | Uses `notificationID`, and wraps `mongo.ErrNoDocuments` so callers can `errors.Is` it. | `DeleteNotification` — `internal/database/notifications.go:104-106` |
| **S7** | The maintenance audit trail was forgeable: `UpdatedBy` came from the request body rather than the verified token. | `UpdatedBy` is now populated from the OIDC `email` claim read out of the request context. | `internal/api/v1/maintenance.go:318` |
| **S8** | `http.MaxBytesReader` was present but its return value was discarded, so the 10 MB body limit did nothing. | The limited reader is captured as `closer` and the JSON decoder reads from it. | `ReadResponse` — `internal/api/httpsuite/response.go:88-90` |
| **X1 / X3** | Prometheus was a **mandatory** startup dependency that was never queried. `PromClient` also built a `RestClient` it never used and called bare `http.Get`. | Removed end-to-end: the `clients/prometheus` package, the config keys, the `CheckURLs` gate, the Helm values/configmap entries and the Compose service are all gone. Metrics now come from the RabbitMQ management API. | deleted — `internal/clients/prometheus/`; `unimq/values.yaml`; `docker-compose.yaml` |
| **X4** | `InsertAlarmEntries` — the correct implementation for appending alarm history — had no callers. | The checker now calls it on both the evaluation-error path and the state-change path. | `internal/notify/checker.go:213,260` |
| **A3** | `NewValidationErrors` ignored the `errors.As` result, so a non-validator error silently produced an empty error map. | The function was deleted; `IsRequestValid` now returns the validator error directly. (The now-orphaned `ValidationErrors` type is tracked as **G28**.) | `internal/api/httpsuite/validation.go:28-34` |
| **R2** | Webhook POSTs used bare `http.Post` with no timeout and no context, so a single hung webhook blocked the checker goroutine and stalled **all** alarm evaluation. | `SendWebhook` builds an `http.NewRequestWithContext` with a 10-second timeout and closes the response body. (Its use of `context.Background()` rather than the caller's context is tracked as **N10**.) | `internal/helpers/notificationhelper/webhook.go:26-60` |
| **H3** | `OIDC_CLIENT_SECRET` was templated into the Helm Secret but never read by the backend — dead config implying a protection that didn't exist. | The secret is bound in config and assigned to the OAuth2 client, so the backend can perform the code exchange. | `internal/config/config.go:48`; `internal/clients/dex/dex.go:32` |
| **E2** | `.env.example` pointed at an external OIDC issuer while the working local setup used the Compose Dex, so copying the example never produced a working environment. | The example now uses `OIDC_URL = "http://localhost:5556/dex"` and the backend callback URL, matching `volumes-for-compose/dex-config-dev.yaml`. | `.env.example:46-53` |
| **G14** | `TestEvaluateMetrics_UnknownType` asserted `require.NoError` against an evaluator that had correctly started returning `unknown rule type`. | The tests were reworked against the new evaluator API and split into `evaluators_test.go`. | `internal/notify/evaluators_test.go` |
| **G15** | Two `httpsuite` tests compared `IsRequestValid`'s result against a `map[string][]string` after its signature changed to `error`, leaving CI red. | Assertions updated to the `error` return. `go test ./...` is green. | `internal/api/httpsuite/validation_test.go` |
| **G16** | `SendEmails` was a method on `*EmailSender` but called the package global `EmailSenderInstance.SendEmail(...)`, ignoring its own receiver, and discarded the `typ mail.ContentType` argument in favour of a hardcoded `"text/plain"`. | It now calls `es.SendEmail(...)` and passes `typ` through. Email code was also split into its own file with per-recipient `EmailStatus` results. | `SendEmails` — `internal/helpers/notificationhelper/email.go:83-91` |
| **`NotifyAlarm` early return** | If `SendWebhooks` failed, the function returned immediately, so **email recipients were skipped whenever any webhook failed**. | It now collects both result sets into a `NotifyStatus` and reports them together via `HasErrors()`. | `NotifyAlarm` — `internal/notify/checker.go:279-294` |
| **R11** *(partial)* | `WriteTimeout` (30s) was shorter than the 60s handler timeout, so the write deadline always fired first and `middleware.Timeout` could never return its 504. | `WriteTimeout` is now 90s, giving the handler timeout room to fire. `ReadHeaderTimeout`/`IdleTimeout` are still unset — this row stays open. | `cmd/unimq/main.go:112-116` |

### Added in the eighth pass

| # | What was wrong | How it was fixed | Where it lives now |
|---|----------------|------------------|--------------------|
| **G18** | `checkMaintenanceSchedules` assigned `SendWebhooks`' `[]WebhookStatus` return to a variable named `err` and then tested `if err != nil`. Because `SendWebhooks` builds its slice with `make(...)`, the result was **never nil**, so the failure branch fired on every run, the success branch was unreachable, and real webhook failures were indistinguishable from success. | The result is captured as `webhookStatus` and iterated, logging per URL on `s.OK` — the same shape `NotifyAlarm` already used. | `checkMaintenanceSchedules` — `internal/notify/checker.go:316-323` |
| **G19** | `ReadyzHandler` wrote the 503 "not ready" response and then fell through into the 200 "ready" response, emitting **two concatenated JSON objects** and a `superfluous WriteHeader` log on every unhealthy probe. | A `return` was added after the 503 branch. | `ReadyzHandler` — `internal/api/v1/health.go:54-57` |

### Added in the ninth pass

| # | What was wrong | How it was fixed | Where it lives now |
|---|----------------|------------------|--------------------|
| **G31** | The G18 reorder moved `SetMaintenanceEntryNotified` after the email block, but the `ErrEmailNotConfigured` branch still did a bare `return` — so the entry was never marked notified **and** the loop aborted, leaving later entries unprocessed and re-sending every webhook on every tick, forever. | The `return` became a `break`, a `notifier.EmailSenderInstance == nil \|\| !IsConfigured()` guard now skips the email block cleanly instead of erroring out of it, and `SetMaintenanceEntryNotified` runs unconditionally at the end of the loop body. | `checkMaintenanceSchedules` — `internal/notify/checker.go:296-348`, marking at `:344` |
| **G20** | `GET /api/v1/rabbitmq/usage` was routed to `GetRMQVhostUsageHandler`, which requires a vhost path parameter the route did not provide, so it returned `400 "vhost name is required"` on every call. It was also absent from the Swagger spec. | The duplicate route was deleted; vhost usage is now reachable only via the vhost-scoped path. Verified: `/rabbitmq/usage` now returns **404**. | Removed from `internal/api/routes/v1/rmq.go` |
| **G21** | The `@Router` annotation on the vhost-usage handler read `/v1/vhost/{vhost-name}/usage` (singular `vhost`), so the documented path 404'd and "Try it out" failed for that endpoint. | The annotation was corrected to `/v1/vhosts/{vhost-name}/usage` and `internal/docs/` regenerated. Verified: `swag init` produces zero diff against the committed spec. | `internal/api/v1/rmq/vhosts.go`; `internal/docs/swagger.json:2185` |
| **G32** | `.env.example:37` promised *"If EMAIL_SMTP_HOST is empty, email sending will be disabled"*, but `sendEmail` only checked `config == nil` and `EmailFromAddress == ""` — never `EmailSMTPHost`. Since `EmailFromAddress` defaults to `unimq@example.com`, "disabled" never triggered and every notification attempted a real SMTP dial to `:25`, once per recipient per tick. | `sendEmail` now consults `config.IsValid()`, which covers the empty-host case, and a new `EmailSender.IsConfigured()` lets callers skip the email block entirely. *(The guard initially landed inverted — see **G35** — and was corrected in `790b321`.)* | `sendEmail` — `internal/helpers/notificationhelper/email.go:39-42`; `IsConfigured` — `:83-93` |

### Added in the tenth pass

Both entries here were **Critical regressions introduced by the ninth-pass fixes** and caught before release. Neither was visible to `go build`, `go vet` or the test suite.

| # | What was wrong | How it was fixed | Where it lives now |
|---|----------------|------------------|--------------------|
| **G34** | The G20/G21 cleanup renamed the chi route parameters to `{vhost-name}`/`{queue-id}` to match the Swagger annotations, but the handlers still called `chi.URLParam(r, "vhost")` / `"queue"`. chi returns `""` for an unknown parameter name, so every handler tripped its own "required" guard — **the whole vhost-scoped RabbitMQ API returned 400**, leaving only `GET /rabbitmq` working. | All seven `chi.URLParam` call sites were updated to `"vhost-name"` / `"queue-id"`. Verified against the real route tree with injected admin claims: all six vhost paths now pass the parameter guard and reach the RabbitMQ client, and `/rabbitmq/usage` still correctly 404s. | `rmq/vhosts.go:69,123,180`; `rmq/queues.go:35,91,111`; `rmq/metrics.go:32` — routes in `internal/api/routes/v1/rmq.go:11-20` (commit `6fc2fe6`) |
| **G35** | The G32 guard landed as `if config.IsValid() { status.Error = ErrEmailNotConfigured; return status }` — the `!` was missing, so the condition was exactly inverted: a **valid** config was rejected and an **invalid** one fell through to the SMTP dial. No alarm or maintenance email could be delivered. | The condition was negated to `if !config.IsValid()`. Verified: a configured sender now proceeds to the dial (`notConfigured=false`), and an empty-host sender returns `SMTP server is not configured` without dialling. | `sendEmail` — `internal/helpers/notificationhelper/email.go:39-42` (commit `790b321`) |
