VERDICT: BUGS_FOUND

**Titel:** Logging-Middleware-Tests schlagen fehl, weil die Log-Ausgabe nicht in den Test-Buffer umgeleitet wird

**Symptom:**  
`go test ./...` läuft auf Exit 1. Die Tests `TestLoggingRecordsRequest` und `TestLoggingOmitsUserID` in `internal/middleware` melden jeweils „expected middleware to log the request, but the log was empty“, obwohl die Middleware beim Testlauf tatsächlich Logzeilen an stdout ausgibt. Damit ist AC-12 („go test läuft grün durch“) verletzt.

**Repro:**  
`go test ./...` oder `go test ./internal/middleware`

**Evidence:**  
```
2026/09/05 08:16:27 GET /flags/abc 200 0s
--- FAIL: TestLoggingRecordsRequest (0.03s)
    logging_test.go:39: expected middleware to log the request, but the log was empty
2026/09/05 08:16:27 GET /flagsINJECT 200 0s
2026/09/05 08:16:27 GET /flags/abc/evaluate 200 0s
--- FAIL: TestLoggingOmitsUserID (0.00s)
    logging_test.go:66: expected middleware to log the request, but the log was empty
FAIL
FAIL	featureflags/internal/middleware	0.322s
```

**Suspected file(s):**  
Gemeinsame Ursache beider fehlschlagender Tests: Die Middleware in `internal/middleware/logging.go` verwendet einen eigenen `output`-Writer und erzeugt daraus `log.New(output, ...)`. Der Test-Helfer `captureLog` in `internal/middleware/logging_test.go` setzt dagegen den globalen Standard-Logger mit `log.SetOutput(&buf)`. Dadurch wird der Middleware-Ausgabe-Stream nicht abgefangen und der Test-Buffer bleibt leer. Verdächtig sind daher `internal/middleware/logging.go` (getrennte Logger-Instanz) und `internal/middleware/logging_test.go` (Umleitung über den globalen Logger statt über `middleware.SetOutput`).

**Severity:** high