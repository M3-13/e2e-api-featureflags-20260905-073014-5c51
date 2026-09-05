VERDICT: BUGS_FOUND

`go build ./...` läuft sauber durch (exit 0). `go test ./...` endet jedoch mit **exit 1**: Das Paket `featureflags/internal/middleware` ist rot. Zwei Tests scheitern an derselben Ursache; zusätzlich bleibt der frühere Race-Nachweis aus.

- **Title:** Logging-Middleware-Tests schlagen fehl, weil der Test den falschen Logger abfängt
  - **Symptom:** `go test ./...` ist rot. `TestLoggingRecordsRequest` und `TestLoggingOmitsUserID` erwarten, dass die Middleware in den umgeleiteten Standard-Logger schreibt; die Middleware nutzt jedoch einen eigenen, fest auf `os.Stdout` verdrahteten Logger. Die Logausgabe erscheint tatsächlich auf stdout, wird von den Tests aber nicht erfasst, sodass die Tests fälschlich melden, das Log sei leer.
  - **Repro:** `go test ./...` ausführen.
  - **Evidence:**
    - `--- FAIL: TestLoggingRecordsRequest (0.04s)`
    - `logging_test.go:39: expected middleware to log the request, but the log was empty`
    - `--- FAIL: TestLoggingOmitsUserID (0.00s)`
    - `logging_test.go:66: expected middleware to log the request, but the log was empty`
  - **Suspected file(s):** `internal/middleware/logging.go` (package-eigener `logger = log.New(os.Stdout, …)`) und `internal/middleware/logging_test.go` (`captureLog` leitet nur den globalen Standard-Logger um, nicht den Middleware-Logger)
  - **Severity:** high

- **Title:** Race-Freiheit weiterhin nicht nachgewiesen – `go test -race` fehlt im Testlauf
  - **Symptom:** Der Report belegt lediglich einen normalen `go test ./...`-Lauf. Der im Sprint versprochene Nachweis paralleler Datenrennenfreiheit (`go test -race ./...`) wird nicht erbracht; AC-12 ist damit nicht belegt.
  - **Repro:** Abschnitt `## Stack: go @ .` ansehen – dort steht nur `### go test ./... (exit 1)`, kein `### go test -race ./...`.
  - **Evidence:** `### go test ./... (exit 1)` — das Flag `-race` erscheint im gesamten Report nicht.
  - **Suspected file(s):** `RUN.json` / Testausführungsvertrag — kein Hinweis, dass der Race-Detektor eingeplant ist.
  - **Severity:** high