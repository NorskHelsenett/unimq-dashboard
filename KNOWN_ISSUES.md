# Known Issues

Findings from a read-only review of the repository. This is a backlog/triage document.

Focus was the **Go backend**; frontend, Helm and dev-environment findings are included where noticed.

Severity: **Critical** (data loss / security breach / core feature broken) · **High** (feature silently wrong) · **Medium** (incorrect behaviour, poor DX, scaling risk) · **Low** (cosmetic, cleanup, docs).

Existing items already tracked in `todos.md` (pagination, alarm cleanup, unique receivers, retry logic, connectivity watchers, more unit tests, Dex redirect to backend) are **not** repeated here except where a concrete defect was found.

Rows are marked ✅ Fixed · ⚠️ Partial/Regressed · ❌ Not fixed. Unmarked rows are not yet addressed.
**The original description and suggested fix are left unchanged for context — see the Re-review section below for current state.**

---

## RE-REVIEW STATUS (baseline `453440c` → `0ca1ca6`, fourth pass)

`go build ./...`, `go vet ./...` and `go test ./...` all pass. Note that **none of the bugs below are caught by the compiler, vet, or the current suite** — they are all runtime behaviour.

**G1 is still open (unchanged this round). Two new Critical regressions (G8, G9) were introduced by the validation and body-limit work.**

A consolidated record of every issue that has been fixed, and how, is at the **bottom of this document**.

### 🚨 Still open — fix these first

| # | What happens now | Where | Impact | Fix |
|---|------------------|-------|--------|-----|
| **G1a** | **Any webhook returning ≥ 400 panics the checker.** The error branch still formats `req.Response.StatusCode` instead of `resp.StatusCode`. `req.Response` is nil on a client request. Unchanged this round. | `internal/helpers/notificationhelper/notificationhelper.go:37` | **Critical.** The one branch that exists to report a bad webhook is the one that crashes, in a goroutine with no `Recoverer`. Reproduced. | `req.Response.StatusCode` → `resp.StatusCode`. |
| **G1b** | **A malformed webhook URL panics the checker.** There is still no `if err != nil` check after `http.NewRequestWithContext`, so `req.Header.Set(...)` runs on a nil `req`. Unchanged this round. | `internal/helpers/notificationhelper/notificationhelper.go:22-23` | **Critical.** Webhook URLs are user-supplied and unvalidated, so any user can permanently kill the checker. Reproduced. | Restore the `if err != nil { lastErr = err; continue }` guard. |
| **G8** | **Adding a notification recipient always fails with 400.** Two bugs compound. (1) `IsRequestValid(r)` is passed the `*http.Request` instead of the decoded `recipient` struct, so it validates `http.Request` — which has no `validate` tags — and never checks the payload. (2) `IsRequestValid` returns `*ValidationErrors`, which is assigned to the `error`-typed `err`. A nil `*ValidationErrors` inside an `error` interface is **not nil**, so `if err != nil` is always true. | `internal/api/v1/notificationrecipient.go:129-138` | **Critical.** `POST` recipient returns 400 "request validation failed" with an empty error body for every request, valid or not. The endpoint is unusable. Reproduced. | Pass the struct: `if ve := httpsuite.IsRequestValid(&recipient); ve != nil { ... }`. Never assign a concrete pointer type into an `error` variable — guard it at the call site, or have `IsRequestValid` return `error` and return a literal `nil`. |
| **G9** | **The request body size limit does nothing.** `http.MaxBytesReader(w, r.Body, 10MB)` is called but its **return value is discarded**. `MaxBytesReader` does not mutate `r.Body`; it returns a new `io.ReadCloser` that must be assigned back. | `internal/routes/httpsuite/response.go:87` | **Critical** as a fix that reads as done but isn't. **S8 is still fully open.** Verified by decoding a 32 MB body through `ReadResponse` with no error and no truncation. | `r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)`, placed **before** the `defer r.Body.Close()`. |

### ⚠️ Worth a second look

- **A2 (validation)** — `validate` tags were added to `PostRecipient` only. `PostAlarmRule` (`models/notification.go:88-95`) still has none, so negative thresholds and empty queue names are still accepted, and no maintenance model is validated. `IsRequestValid` is called from exactly one handler — and that call is **G8**.
- **A3** — unchanged. `NewValidationErrors` still ignores the `errors.As` result, which is why G8 surfaces as an empty error map rather than a useful message.
- **R7** — unchanged. `writeJSONResponse` still logs a marshalling failure and then writes the nil body with a success status. Only the recursion (R6) was removed.
- **S11 (`https://` defaults)** — `RabbitMQHost` and `PrometheusHost` now default to `https://localhost`, but **`.env.example` still says `http://localhost`** (lines 20 and 33), and both services speak plain HTTP on their default ports (RabbitMQ management 15672, Prometheus 9090). Anyone relying on the built-in defaults now fails to connect, and since Prometheus connectivity is a hard startup requirement (`CheckURLs`), **the service will not boot**. Either enforce TLS explicitly with a clear error, or keep the `http://` default and reconcile `.env.example` — see **E1**/**E3**.
- **B3 / B8 / R11** — unchanged from the previous pass; see the fixed log and the backlog tables below.

### Not yet addressed

B5, B14–B17, B19, B21–B23 · S2–S5, S7–S10, S12 (S8 attempted, see G9) · R3–R5, R7 · D1–D5 · X1–X8 · A1, A3–A10, A12 · N1–N9 · O1–O5 — plus all FRONTEND, HELM, DOCKER, CI/CD and DEV-ENV items. Note **H4** (TCP-only probes) can now be closed out in the chart, since A11 gave it real endpoints to probe.

---

## BACKEND

### Correctness / logic bugs

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| B1 ✅ Fixed | Alarm history only ever records the **first** event per rule | `internal/models/alarms.go:10` (`AlarmID` is `bson:"_id"`), `internal/database/alarms.go:57` `AddAlarm` | `AddAlarm` does `InsertOne` on every fired/resolved transition using the rule ID as `_id`. The second transition fails with a duplicate-key error, which is only logged. Alarm history is permanently incomplete. | **Critical** | Use the already-written-but-never-called `InsertAlarmEntries` (`$push` into `entries`), or upsert. |
| B2 ✅ Fixed | Maintenance notifications only reach the **first** vhost's recipients | `internal/notify/checker.go:161` | `checkMaintenanceSchedules` is called *inside* the per-vhost loop but marks entries `Notified=true` globally on the first pass. Every other vhost's recipients never get maintenance notices. | **High** | Call it once per tick, outside the vhost loop, with the union of all webhook URLs. |
| B3 ⚠️ Partial | Email recipients **never** receive real alarms | `internal/notify/checker.go:159,286` | The checker only calls `WebhookURLs()` + `SendWebhooks`. `EmailRecipients()`/`SendEmail` are used *only* by the test endpoint (`notificationrule.go:540`). Users configure email recipients and silently get nothing. | **High** | Send to both webhook and email recipients in `Notify`. |
| B4 ✅ Fixed | `GetMaintenanceHandler` writes two responses on history failure | `internal/api/v1/maintenance.go:49-60` | On error it calls `WriteJSONError` (500 + body) and then **falls through** to `SendResponse` (200 + body). Produces `superfluous WriteHeader` and a corrupt double-JSON body. | **High** | Add the missing `return`, or drop the error write and serve partial data. |
| B5 | Timezone handling is inconsistent for maintenance windows | `internal/models/maintenance.go:70,124` (`ParseInLocation`/`time.Local`) vs `internal/api/v1/maintenance.go:268` (`time.Parse` = UTC) | Creating an entry parses in server-local time; editing it parses the same layout as UTC. Editing a window silently shifts it by the UTC offset. | **High** | Pick one (UTC) and use a single shared parse helper; ideally accept RFC3339. |
| B6 ✅ Fixed | Configured RabbitMQ limits are silently discarded | `internal/clients/rabbitmq/rabbitmq.go:113` | `NewRMQClient` returns `&RMQClient{restClient: restclient}` and never assigns `Limits`. `WithRMQLimits` and `RABBITMQ_*_LIMIT` are dead. `APIService.RMQLimits` is likewise always `nil` (no option sets it). | **High** | Set `Limits: config.Limits` in the constructor; add `WithRMQLimits` to `APIService` or delete the field. |
| B7 ✅ Fixed | Mongo URI is malformed when `MONGODB_HOST` includes a scheme | `internal/database/database.go:124` | `CreateUri` runs `url.QueryEscape(host)`. The Go default is `"mongodb://localhost"` (`config.go:64`), which escapes to `mongodb%3A%2F%2Flocalhost`, producing `mongodb://mongodb%3A%2F%2Flocalhost:27017`. `.env` avoids this by setting a bare host, so the bug only bites when the var is unset. | **High** | Make the default a bare host, strip any scheme, and use `url.UserPassword`/`PathEscape` rather than `QueryEscape`. |
| B8 ❌ Not fixed | Missing queues evaluate to `0` instead of erroring | `internal/notify/checker.go:328-357` | For queue-scoped rule types, if no queue matches `rule.QueueName` the loop exits and `v` stays `0`, then `v >= rule.Threshold` is returned. With `threshold <= 0` this **fires a false alarm** for a deleted/renamed queue. | **High** | Return a distinct "target not found" error and skip/flag the rule. |
| B9 ✅ Fixed | Rule update wipes fields the client didn't send | `internal/api/v1/notificationrule.go:300-320` | `AlarmRuleUpdate` is decoded into a zero-valued struct, then threshold *and* message are always written. Omitting `message` blanks it; omitting `threshold` sets it to `0`. The two writes are also non-atomic — a failure after the first leaves a partial update. | **High** | Use pointer fields (or `$set` only present keys) and combine into a single `UpdateOne`. |
| B10 ✅ Fixed | Non-`ErrNoDocuments` DB errors are swallowed | `internal/api/v1/internal.go:34-50` | In `ensureNotificationHostExists`, the `if errors.Is(err, mongo.ErrNoDocuments)` block has no `else` and the outer `if err != nil` never returns. A real DB failure falls through to a second query. | **Medium** | Return the error when it isn't `ErrNoDocuments`. |
| B11 ✅ Fixed | `LOG_LEVEL` can never be set *less* verbose than Info | `internal/logger/logger.go:34-36` | `UpdateLogLevel` returns early if `slog.Default().Enabled(ctx, level)` is true. Default is Info, so Warn/Error are already "enabled" and the update is skipped. | **Medium** | Always rebuild the handler; drop the guard. |
| B12 ✅ Fixed | `request_id` is never attached to logs | `internal/logger/logger.go:15` | Looks up `ctx.Value("request_id")` (raw string key), but chi's `middleware.RequestID` stores under its own unexported key type. Always a miss — dead code. | **Medium** | Read `middleware.GetReqID(ctx)`. |
| B13 ✅ Fixed | `middleware.RequestID` registered *after* `middleware.Logger` | `internal/routes/routes.go:50-51` | Logger runs first, so request IDs are absent from its output even once B12 is fixed. | **Low** | Register `RequestID` (and `RealIP`) before `Logger`. |
| B14 | `MaintenanceEntry.UnmarshalJSON` validates `status` but never assigns it, and force-resets `Notified` | `internal/models/maintenance.go:105-145` | Decoded entries always get an empty `Status` and `Notified=false`, silently discarding state. | **Medium** | Assign `e.Status = ParseMaintenanceStatus(aux.Status)`; don't reset `Notified`. |
| B15 | No validation that maintenance `End > Start` | `internal/models/maintenance.go:67-88,166-181` | An inverted window is accepted; `AdvanceMaintenanceStatuses` then immediately flips it to `done`. | **Medium** | Reject `End <= Start` in both `ToMaintenanceEntry` and `Validate`. |
| B16 | `CheckVhostExists` decodes notification docs into `[]AlarmEntry` | `internal/database/vhost.go:36-60` | Wrong target type (copy-paste). Works only because it just checks `len() > 0`. Fragile and misleading. | **Medium** | Use `CountDocuments` with the `_id` filter. |
| B17 | `notified` is stored per-vhost, not per-rule | `internal/database/notificationrules.go:117` | `UpdateNotificationRule` sets a top-level `notified` field, so the last rule evaluated overwrites the flag for every other rule on the vhost. | **Medium** | Move to `rules.$.notified`. |
| B18 ✅ Fixed | Writes `rules.$.lastResolved`, a field the model doesn't have | `internal/database/notificationrules.go:124` | Written on every resolve but never read or returned — schema drift. | **Low** | Add `LastResolved *time.Time` to `AlarmRule` or drop the write. |
| B19 | `updated_at` stored as a string while `start`/`end` are dates | `internal/database/maintenance.go:196` vs `models/maintenance.go:98` | Mixed types in one collection; range queries and sorting on `updated_at` won't work. `MaintenanceEditLog.UpdatedAt` is correctly a `time.Time`, adding to the inconsistency. | **Medium** | Store `time.Time`. |
| B20 | Error message interpolates the constant `_id` instead of the vhost | `internal/database/notifications.go:99` | Produces `notification not found for vhost _id.` | **Low** | Use `notificationID`. |
| B21 | `isPresent` never rejects an int, and panics on other types | `internal/config/config.go:250-263` | `case int: return true` means `BASE_PORT=0` etc. always pass validation. The `default` branch `panic`s. | **Medium** | Validate ranges explicitly; return `false` instead of panicking. |
| B22 | `.env` load errors are silently discarded | `internal/config/config.go:100,161-167` | `errors.Is(viper.ConfigFileNotFoundError{}, err)` has its arguments reversed *and* `SetConfigFile` returns `*fs.PathError` for a missing file anyway. The caller then discards the result with `_ =`. A malformed `.env` is invisible. | **Medium** | `errors.As` on the right operand and surface non-"not found" errors. |
| B23 | `ADMIN_GROUPS` is not validated as required | `internal/config/config.go:218-240` | Defaults to empty, which makes `IsAGroupInClaim` fail for everyone — the whole API returns 403 with no obvious cause. Fails closed (good) but is undiagnosable. | **Medium** | Add to `validateConfiguration` or log a loud warning at startup. |

### Security

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| S1 ✅ Fixed | Unchecked type assertion on the `groups` claim → panic | `internal/routes/httpsuite/context.go:57` | `aClaimGroups.([]any)` panics if an authenticated token carries `groups` as a string or any non-array. Caught by `Recoverer` → 500, but it's a remote-triggerable panic on every endpoint. | **High** | Use the comma-ok form and return `ErrInvalidGroupsType` (already defined, currently unused). |
| S2 | Internal error text is returned to clients | `internal/routes/httpsuite/error.go:73-77`, used with `WithError(err)` throughout `api/v1` | `ErrorResponse.Error` carries `err.Error()`. Combined with S3 this leaks raw upstream RabbitMQ/Mongo messages, hostnames and driver internals to any caller. | **Medium** | Log the error; return only `externalErrorMessage`. Gate `WithError` behind debug mode. |
| S3 | Raw upstream response body becomes the error | `internal/clients/rest/rest.go:172` | `errors.New(string(bodyBytes))` turns the entire RabbitMQ error page into a Go error that is then surfaced via S2. | **Medium** | Wrap in a typed error with a truncated, sanitised message. |
| S4 | Webhook URLs are user-supplied and unvalidated → SSRF | `internal/helpers/notificationhelper/notificationhelper.go:12-30`, recipient creation in `api/v1/notificationrecipient.go` | The server POSTs to arbitrary URLs supplied through the API, including internal addresses and cloud metadata endpoints. The `/rules/{rule}/test` endpoint makes this trivially on-demand. | **High** | Validate scheme/host, enforce an allowlist or deny RFC1918/link-local, and rate-limit the test endpoint. |
| S5 | Authorization is all-or-nothing; ACL layer is an empty stub | `internal/api/v1/acl.go` and `internal/database/acl.go` are 1-line package declarations; every handler only calls `IsAGroupInClaim(ctx, rc.AdminGroups)` | There is no per-vhost authorization. Any member of any admin group can read and mutate **every** vhost, rule, recipient and maintenance window. | **High** | Implement the ACL collection (already provisioned in `initCollections` and `mongo-init.js`) and enforce per-vhost scope. |
| S6 ✅ Fixed | `AuthorizeScope` comparison is inverted | `internal/models/acl.go:30-34` | `ScopeAdmin = 0` is the *highest* privilege, but the check is `allowed <= scope`. `AuthorizeScope(ScopeRead, ScopeAdmin)` → `0 <= 2` → **true**, granting a read-only user admin. Currently unreachable (S5) but a landmine. | **High** (latent) | Invert to `scope <= s`, add unit tests, and type the constants as `Scope`. |
| S7 | Maintenance audit trail is client-controlled | `internal/api/v1/maintenance.go:255,289` | `UpdatedBy` comes from the request body, not the verified OIDC claims. The audit log is forgeable. | **Medium** | Derive `updated_by` from the token's `email`/`sub` claim. |
| S8 ❌ Not fixed (attempted) | No request body size limit | `internal/routes/httpsuite/response.go:86-99` | `ReadResponse` decodes straight from `r.Body`. An arbitrarily large body can exhaust memory. | **Medium** | Wrap with `http.MaxBytesReader`. |
| S9 | Swagger UI is unauthenticated | `internal/routes/v1.go:9-13` | `/api/swagger/*` is registered in the unprotected group and documents the full API surface. | **Low** | Move behind auth or disable in production builds. |
| S10 | No CORS, rate limiting, or security headers | `internal/routes/routes.go:50-53` | Only `Logger`, `RequestID`, `Recoverer`, `Timeout` are registered. | **Medium** | Add `cors`, `httprate`, and standard security headers. |
| S11 ⚠️ Partial | RabbitMQ basic-auth credentials may go over plaintext HTTP | `internal/config/config.go:66` default `http://localhost`; no TLS enforcement in `rest.go` | Credentials are sent in a `Basic` header; with an `http://` host they're in cleartext. | **Medium** | Require `https://` outside local dev; make TLS config explicit. |
| S12 | `gosec` suppressed on the outbound request | `internal/clients/rest/rest.go:158` — `//nolint:gosec // if this causes an exploitation there are bigger issues` | Silences the variable-URL (SSRF) warning that S4 then demonstrates is real. | **Low** | Remove the suppression once URLs are validated. |

### Resource leaks & HTTP layer

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| R1 ✅ Fixed | Response body leaked on every non-2xx upstream response | `internal/clients/rest/rest.go:167-175` | The `StatusCode > 399` branch reads the body and returns without `Close()`. Same on the `io.ReadAll` error path and the JSON-decode error path in `handleOutput`. Leaks a connection per RabbitMQ error. | **High** | `defer resp.Body.Close()` immediately after `Do`. |
| R2 ⚠️ Partial | Webhook POSTs have no timeout and no context | `internal/helpers/notificationhelper/notificationhelper.go:18` | Bare `http.Post` uses the default client (no timeout). One hung webhook blocks the checker goroutine indefinitely, stalling **all** alarm evaluation. | **High** | Use a client with a timeout and `http.NewRequestWithContext`. |
| R3 | `CheckURLs` opens three TCP connections and never closes them | `internal/config/config.go:110-135` | The `net.Conn` returns are discarded with `_`. Minor, but they stay open for the process lifetime. | **Low** | Assign and `Close()` each. |
| R4 | REST client pins the process-lifetime context onto every request | `internal/clients/rest/rest.go:131,137` | `http.NewRequestWithContext(r.Context, ...)` uses the startup context, so per-request cancellation and deadlines never propagate to RabbitMQ calls. | **Medium** | Pass the caller's `ctx` through each method. |
| R5 | In-memory queue history grows unboundedly | `internal/clients/rabbitmq/rabbitmq.go:26-41` | Package-level global `map[string][]int` keyed by `vhost/queue`, never evicted. Deleted queues leak forever. Already flagged by a `TODO` in the file. | **Medium** | Persist to Mongo or add TTL eviction; move `historySize` to config. |
| R6 ✅ Fixed | `writeJSONResponse` can recurse infinitely on write failure | `internal/routes/httpsuite/response.go:73-84` | The error path calls itself, which writes again, fails again, recurses again. Also calls `WriteHeader` a second time. | **Medium** | Log and return; never re-enter. |
| R7 | `writeJSONResponse` writes a nil body when marshalling fails | `internal/routes/httpsuite/response.go:75-78` | Logs the error but continues and writes `nil`, sending a success status with an empty body. | **Medium** | Return 500 and stop. |
| R8 ✅ Fixed | Mongo client is never disconnected on shutdown | `cmd/unimq/main.go` | No `client.Disconnect()` in the shutdown path. | **Low** | Disconnect after `wg.Wait()`. |
| R9 ✅ Fixed | Double `WaitGroup` accounting | `cmd/unimq/main.go:110-124` | `wg.Add(1)` **and** `wg.Go(...)` (which adds internally) **and** `defer wg.Done()` inside. Adds 2 / Dones 2 — balances only by accident. | **Low** | Use `wg.Go(f)` alone. |
| R10 ✅ Fixed | Data race on the shared `err` variable | `cmd/unimq/main.go:115,144` | The server goroutine assigns `err = server.ListenAndServe()` while main later assigns `err = server.Shutdown(...)`. | **Medium** | Use a goroutine-local variable or an error channel. |
| R11 ⚠️ Partial (timeout race fixed) | `WriteTimeout` (30s) is shorter than the handler timeout (60s) | `cmd/unimq/main.go:106`, `internal/routes/routes.go:53` | The write deadline always fires first, so `middleware.Timeout` never produces its 504. No `ReadHeaderTimeout`/`IdleTimeout` either (Slowloris exposure). | **Medium** | Align the values and set the missing timeouts. |

### MongoDB data layer

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| D1 | `UpdateOne` results are ignored almost everywhere | `notificationrules.go` (add/delete/toggle/threshold/message), `maintenance.go:110,169` | `MatchedCount == 0` is treated as success, so deleting or updating a non-existent rule/entry returns **200**. Only `PatchMaintenanceEntry`, `DeleteMaintenanceEntry` and `DeleteNotificationRecipient` check. | **Medium** | Check `MatchedCount`/`ModifiedCount` and return a not-found error consistently. |
| D2 | Not-found surfaces as **500** on several endpoints | `maintenance.go:99` (expects `ErrMaintenanceNotFound`, gets raw `mongo.ErrNoDocuments`), `notificationrecipient.go:62,212`, `notification.go:150`, `notificationrule.go:383` | Clients cannot distinguish "missing" from "broken". | **Medium** | Map `mongo.ErrNoDocuments` to a sentinel in the DB layer and `errors.Is` it in handlers. |
| D3 | No indexes on any collection | `internal/database/database.go:145-155` | `alarms`, `maintenance_edit_logs.maintenance_id` and all filters are collection scans. | **Medium** | Create indexes at startup. |
| D4 | No pagination or limits on list endpoints | `GetAlarmsAll`, `GetMaintenanceAll`, `GetNotificationsAll`, `GetMaintenanceEditLogs` | Unbounded result sets grow forever (alarms especially). Already noted in `todos.md`. | **Medium** | Add `limit`/`skip`/date filters. |
| D5 | `UpdateNotification` `$set`s the whole struct including `_id` | `internal/database/notifications.go:74-79` | MongoDB rejects mutating the immutable `_id`. Function appears unused — likely broken if ever called. | **Low** | Exclude `_id` from the update document, or delete the function. |

### Dead code / unused features

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| X1 | Prometheus is a **mandatory** dependency that is never used | `internal/clients/prometheus/prometheus.go` (`QueryRange` has no callers), `config.go:236-237` (required), `config.go:119-124` (`CheckURLs` hard-fails) | The app refuses to start unless Prometheus is reachable, yet never queries it. | **High** | Either wire `QueryRange` into a metrics endpoint or make Prometheus optional/remove it. |
| X2 | `CheckURLs` makes every dependency a hard startup gate | `internal/config/config.go:108-138` | A momentary RabbitMQ/Mongo/Prometheus blip at boot kills the process. In Kubernetes this becomes a `CrashLoopBackOff` instead of an unready pod. | **Medium** | Warn and retry with backoff; expose readiness via a health endpoint instead. |
| X3 | `PromClient` builds a `RestClient` it never uses | `prometheus.go:28-39,58` | `QueryRange` calls bare `http.Get`, bypassing auth, timeout and context. The `username`/`password` params are ignored (and passed as `""` from `routes.go:24`). | **Medium** | Route the query through `RestClient`. |
| X4 | `InsertAlarmEntries` and `DeleteAlarm` have no callers | `internal/database/alarms.go:69,80` | `InsertAlarmEntries` is the correct implementation for B1 but is never invoked. | **Medium** | Use it (see B1) or delete. |
| X5 | `rest.WithUsername` / `WithPassword` are no-ops | `internal/clients/rest/rest.go:74-84` | The `Config` fields are set but never read; auth comes solely from the auth provider. | **Low** | Remove the options. |
| X6 | `AlarmRule.IsTriggered` duplicates `evaluate`'s comparison and is unused | `internal/models/notification.go:146` | Two sources of truth for the threshold rule. | **Low** | Delete or make `evaluate` call it. |
| X7 | Unused sentinel errors and dead branches | `httpsuite/context.go:17-18` (`ErrGroupsNotFound`, `ErrInvalidGroupsType` unused); `database/notificationrules.go:31-35` and `api/v1/internal.go:36-41` (identical if/else branches) | Suggests incomplete error handling that was never finished. | **Low** | Wire them up or remove. |
| X8 | `AlarmStatus` has six values; only two are ever used | `internal/models/notification.go:167-175` | Only `ok` and `firing` are written; `active` is the initial value, `fired`/`inactive`/`unknown` are dead. The state machine is unclear. | **Medium** | Reduce to the states actually used and document transitions. |

### API design & consistency

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| A1 | Response models lack `json` tags → PascalCase keys | `models/notification.go:268` (`VhostNotification`), `models/alarms.go:9` (`AlarmEntry`), `models/maintenance.go:200,209` (`MaintenanceAdminResponse`, `MaintenanceResponse`) | These are returned directly from handlers, so the API emits `Name`/`Recipients`/`Scheduled`/`History`/`AlarmID` while the rest of the API is `snake_case`. Forces the frontend to match two conventions. | **Medium** | Add `json` tags throughout. |
| A2 ⚠️ Partial/Regressed | No struct validation despite a validator being available | `models/notification.go:17-22,84-91`; `httpsuite.IsRequestValid` exists but is **never called** | `PostRecipient` accepts an empty/garbage URL for a webhook and an invalid email address; `PostAlarmRule` accepts negative thresholds and an empty queue name for queue-scoped rules. | **Medium** | Add `validate:"required,url,email"` tags and call `IsRequestValid` in handlers. |
| A3 | `NewValidationErrors` ignores the `errors.As` result | `httpsuite/validation.go:15-17` | If the error isn't a `validator.ValidationErrors`, it silently returns an empty error map, so the caller reports "invalid" with no detail. | **Medium** | Check the boolean and handle the non-validation case. |
| A4 | Double URL-unescaping of path parameters | every handler, e.g. `api/v1/vhosts.go:70`, `queue.go:43` | `net/http` already decodes `r.URL.Path`, so `chi.URLParam` returns a decoded value. `url.QueryUnescape` decodes a second time and also converts `+` to a space — mangling vhost/queue names containing `+` or `%`. | **Medium** | Drop the extra unescape; use `chi.URLParam` directly. |
| A5 | ~40 lines of identical boilerplate repeated in 13+ handlers | all of `internal/api/v1/*.go` | The group check / param extraction / unescape preamble is copy-pasted, which is exactly how the inconsistencies in D2 and A6 crept in. | **Medium** | Extract an authorization middleware and a param-decoding helper. |
| A6 | Inconsistent status codes for the same class of failure | `maintenance.go:145` (bad time format → **500**, should be 400), `vhosts.go:75` and `notificationrecipient.go:44` (decode failure → 500 vs 400 elsewhere) | Client error handling can't be uniform. | **Medium** | Standardise: 400 for input, 404 for missing, 500 for internal. |
| A7 | `WithError(err)` passed where `err` is provably `nil` | `queue.go:110`, `notificationrule.go:432,523`, `notificationrecipient.go:207` | Copy-paste; produces a misleading empty `error` field. | **Low** | Remove those options. |
| A8 | Test-notification endpoint rejects email-only vhosts | `api/v1/notificationrule.go:520-528` | Returns **400** when there are no webhook URLs, even if email recipients exist, so email config can't be tested. It also aborts mid-loop after partially sending. | **Medium** | Fail only when *no* recipients of any kind exist; collect per-recipient results. |
| A9 | Rule updates use `POST` instead of `PUT`/`PATCH` | `internal/routes/v1.go:55-57` | `POST /rules/{rule}` for an update is non-RESTful and collides conceptually with rule creation. | **Low** | Use `PUT`/`PATCH`. |
| A10 | Swagger annotations don't match the routes | path params documented as `{vhost-name}`/`{queue-id}`/`{rule-id}` but registered as `{vhost}`/`{queue}`/`{rule}`; `status.go:17` documents `/v1/checker/status` but the route is `/v1/status`; `maintenance.go` Patch/Logs handlers omit `@security bearer`; `notificationrule.go:257` has `[Post]` capitalised | Generated docs and clients are wrong. | **Medium** | Align annotations with `internal/routes/v1.go` and regenerate. |
| A11 ✅ Fixed | No health/readiness endpoint | `internal/routes/v1.go` | The only status endpoint is `/api/v1/status`, which is **behind auth** — unusable for Kubernetes probes (see H6). | **Medium** | Add unauthenticated `/healthz` and `/readyz`. |
| A12 | Duplicate recipients are accepted | `api/v1/notificationrecipient.go:138-146` | No uniqueness check, so the same webhook can be added N times and receives N copies of every alarm. Listed in `todos.md`. | **Low** | Deduplicate on URL/email before `$push`. |

### Notification checker specifics

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| N1 | State diverges when the status update fails | `internal/notify/checker.go:257-262` | If `UpdateNotificationRule` errors, the code only logs and continues to log the alarm and notify. The DB still says `ok`, so the alarm re-fires and re-notifies every interval. | **Medium** | Return early on update failure. |
| N2 | Only `>=` comparisons are supported | `internal/notify/checker.go:357`, `models/notification.go:146` | Rules cannot express "below threshold" (e.g. consumer count dropped to 0, throughput stalled). | **Medium** | Add a comparison operator to `AlarmRule`. |
| N3 | Checker interval is hardcoded | `cmd/unimq/main.go:88` | `60*time.Second` is not configurable via env, unlike every other tunable. | **Low** | Add `CHECKER_INTERVAL_S`. |
| N4 | `GetMetrics` fetches **all** cluster connections and channels per vhost | `internal/clients/rabbitmq/rabbitmq.go:224-241` | Two full cluster-wide listings per vhost per tick, filtered client-side. With many vhosts this is O(vhosts × cluster size) every 60s and will hammer the management API. | **Medium** | Use `/vhosts/{vhost}/connections` / `/channels`, or fetch once per tick and group. |
| N5 | A rule type is signalled as an error | `internal/notify/checker.go:180-182` | `AlarmTypeMaintenance` returns `ErrNotificationRuleInMaintenance` from `EvaluateMetrics`, conflating "not applicable" with failure. | **Low** | Skip it explicitly rather than via an error. |
| N6 | One-shot startup delay uses a `Ticker` | `internal/notify/checker.go:86` | A `Ticker` is created for a single 15s wait. | **Low** | Use `time.NewTimer`/`time.After`. |
| N7 | `SendEmail` ignores port, username and password | `helpers/notificationhelper/notificationhelper.go:47` | `mail.NewClient(config.EmailSMTPHost)` only — `EMAIL_SMTP_PORT/USERNAME/PASSWORD` are dropped, so authenticated or non-default-port SMTP fails. A new client is also constructed per message. The guard checks `EmailFromAddress` but reports "SMTP server is not configured". | **Medium** | Reuse the already-configured `APIService.EmailClient`. |
| N8 | Webhook payload is Slack-shaped and hardcoded | `helpers/notificationhelper/notificationhelper.go:13-14` | `{"text": ...}` works for Slack/Teams but not generic webhooks; the `json.Marshal` error is discarded, and response bodies aren't drained before close. | **Low** | Make the payload template configurable; handle the error. |
| N9 | `BuildMessage` switches on string literals, not the `AlarmType` constants | `models/notification.go:220-239` | Renaming a constant silently falls through to the generic message. | **Low** | Switch on the typed constants. |

### Observability & testing

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| O1 | Test coverage is minimal | tests exist only in `internal/database`, `internal/notify`, `internal/routes/httpsuite` | No tests for `api/v1` (the bulk of the logic), `config`, `clients/*`, or `logger`. Most bugs above are trivially testable. Already in `todos.md`. | **Medium** | Add handler tests with a mock RMQ/DB and table-driven `evaluate` tests. |
| O2 | Logs are `TextHandler`, not JSON | `internal/logger/logger.go:29` | Harder to parse in Kubernetes log aggregation. | **Low** | Use `slog.NewJSONHandler` (optionally env-switched). |
| O3 | Copy-pasted, incorrect log statement | `clients/prometheus/prometheus.go:63` | Logs `"failed to close cursor"` (no cursors here) and `time.Since(start)` where `start` is the *query window* start, not a timer. | **Low** | Fix the message and use a real timer. |
| O4 | Unchecked assertion in the log handler | `internal/logger/logger.go:56` | `a.Value.Any().(*slog.Source)` will panic if the attr type ever differs. | **Low** | Use the comma-ok form. |
| O5 | `slog.Error` used for an informational message | `internal/logger/logger.go:38` | "updating log level" is logged at Error. | **Low** | Use `slog.Info`. |

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
| F11 | `NotifyRule` loads one vhost and navigates back to another | `web-src/src/pages/NotifyRule.tsx:10-20` | Loads with `params.get('vhost') || selected` but the back link always uses `selected`. | **Medium** | Derive a single validated vhost value. |
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
| H3 | `OIDC_CLIENT_SECRET` is templated but never read by the backend | `unimq/templates/secret.yaml:13` vs `internal/config/config.go:44-48` | Dead configuration that implies a protection the backend doesn't provide — and relates directly to F1. | **High** | Implement the backend code exchange (per `todos.md`) or remove the secret. |
| H4 | Probes are TCP-only | `unimq/templates/backend-deployment.yaml:52-61`, `frontend-deployment.yaml:61-70` | The backend can accept TCP while RabbitMQ/Mongo are unreachable, so a broken pod stays in service. Compounded by A11 (no HTTP health endpoint exists to probe). | **Medium** | Add `/healthz`/`/readyz` and switch to `httpGet`. |
| H5 | One `replicaCount` drives both deployments, defaulting to 1 | `unimq/values.yaml:4`, both deployment templates line 10 | Backend and frontend have different scaling needs; a single backend replica is a SPOF. Combined with X2, a dependency blip takes the whole service down. | **Medium** | Split the values; add optional HPA/PDB. |
| H6 | Resource requests/limits are hardcoded | `backend-deployment.yaml:40-44`, `frontend-deployment.yaml:45-49` | Cannot be tuned per environment; risks throttling or OOM kills. | **Low** | Move into `values.yaml`. |
| H7 | Namespace is hardcoded and not created by the chart | `unimq/values.yaml:3` and all templates | Install fails unless the namespace is pre-created. | **Low** | Document the prerequisite or template it optionally. |
| H8 | No `_helpers.tpl`, no standard labels/selectors | `unimq/templates/` | Missing the conventional `app.kubernetes.io/*` labels and name/fullname helpers makes selectors brittle. | **Low** | Add a helpers template and standard label blocks. |

---

## DOCKER / CONTAINERS

| # | Issue | Where | Why it matters | Severity | Potential fix |
|---|-------|-------|----------------|----------|---------------|
| C1 | Compose ships default credentials | `docker-compose.yaml:24-25,49,68-71` | RabbitMQ `guest/guest`, MongoDB `admin/adminpassword`, Mongo Express `test/test`. | **High** (with C2) | Source from an ignored `.env`; fail if absent. |
| C2 | All dev services are published on every host interface | `docker-compose.yaml:6,17,31,45,64` | Dex, RabbitMQ, Prometheus, MongoDB and Mongo Express are reachable from the local network, with the credentials in C1. | **High** | Bind to `127.0.0.1`. |
| C3 | Frontend build output path is confusing and fragile | `web-src/vite.config.ts:13` (`outDir: "../web-src/static/dist"`) and `dockerfiles/Dockerfile.frontend:12` (`COPY --from=builder /web-src/static/dist`) | The `../web-src/` prefix means the output lands in `web-src/static/dist` locally but at the container **root** `/web-src/static/dist` in the builder (where `WORKDIR` is `/app`). It works only by coincidence of the two mistakes cancelling out. | **Medium** | Set `outDir: "static/dist"` and copy from `/app/static/dist`. |
| C4 | Mongo Express connection URL has a stray trailing quote | `docker-compose.yaml:70` | `ME_CONFIG_MONGODB_URL` ends with `unimq"`, likely breaking the connection. | **Medium** | Remove the quote. |
| C5 | Base images are not pinned | `Dockerfile.frontend:3` (`node:22-alpine`), `Dockerfile.backend:17` (`gcr.io/distroless/static`, no tag) | Rebuilds are not reproducible and can silently pull new vulnerabilities. | **Medium** | Pin to digests or exact patch versions. |
| C6 | Unused `GCR_MIRROR` build arg | `Dockerfile.frontend:1` | Declared but never referenced. | **Low** | Use it or remove it. |
| C7 | Mongo Express doesn't wait for MongoDB health | `docker-compose.yaml:73-74` | Uses plain `depends_on` despite MongoDB having a healthcheck, so it can start too early and fail. | **Low** | Use `condition: service_healthy`. |
| C8 | Only MongoDB has a healthcheck | `docker-compose.yaml` | Startup ordering and local debugging are unreliable for Dex, RabbitMQ and Prometheus. | **Low** | Add healthchecks. |
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
| E1 | `.env.example` documents variables the backend doesn't read | `.env.example:23-25` — `RABBITMQ_CONNECTION_MAX`, `RABBITMQ_QUEUE_MAX`, `RABBITMQ_MESSAGE_UNACK_LIMIT` | None are bound in `internal/config/config.go`; the real names are `RABBITMQ_CONNECTION_LIMIT` / `RABBITMQ_QUEUE_LIMIT`. Developers will set values that do nothing — and per B6 even the correct ones are discarded. | **Medium** | Reconcile the example with `loadEnvironmentVariables`. |
| E2 | `.env.example` doesn't match the documented local setup | `.env.example:39-41` vs `.env:39-42` | The example points at an external OIDC issuer while the working local config uses the Compose Dex instance. Copying the example doesn't produce a working environment. | **Medium** | Make the example reflect the local Docker/Dex setup; add a separate production sample. |
| E3 | `MONGODB_HOST` semantics are ambiguous | `.env:12` (`"localhost"`) vs `config.go:64` default (`"mongodb://localhost"`) | The two disagree on whether a scheme is expected; `CheckURLs` strips it, `CreateUri` escapes it. This is the root of B7. | **Medium** | Document and enforce one format. |
| E4 | Local `.env` uses predictable credentials | `.env:11-20,24-25` | Low risk on its own, but dangerous combined with C2 (services bound to all interfaces). Confirmed **not committed** — `git check-ignore` reports `.gitignore:5` and `git ls-files .env` is empty. | **Low** | Generate local-only credentials. |

### Verified as NOT issues

- **No secrets or database files are committed.** `git ls-files volumes-for-compose` returns only three config files (`dex-config-dev.yaml`, `mongodb/mongo-init.js`, `prometheus-config.yaml`). No MongoDB data directory is tracked.
- **`.env` is correctly gitignored** and absent from the index.
- `go build ./...` and `go vet ./...` both pass cleanly.

---

## FIXED — CHANGE LOG

Everything below has been verified fixed in the current tree (`2afca93`). Locations are where the fix lives **today**.

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
