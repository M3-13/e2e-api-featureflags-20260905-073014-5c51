VERDICT: BUGS_FOUND

**Bug: Logging-Middleware-Tests schlagen fehl – Logausgabe wird nicht abgefangen**

- **Titel**: Logging-Middleware-Ausgabe umgeht den im Test gesetzten Logger
- **Symptom**: `go test ./...` läuft nicht grün durch. Zwei Middleware-Tests brechen ab: Sie erwarten, dass die Logging-Middleware in den von den Tests bereitgestellten Puffer schreibt, erhalten aber einen leeren Puffer. Die Logzeilen erscheinen stattdessen auf stdout, nicht im Test-Buffer.
- **Repro**: `go test ./...` im Projekt ausführen. Konkret scheitern:
  - `TestLoggingRecordsRequest`
  - `TestLoggingOmitsUserID`
- **Evidence**:
  ```
  --- FAIL: TestLoggingRecordsRequest (0.02s)
      logging_test.go:39: expected middleware to log the request, but the log was empty
  --- FAIL: TestLoggingOmitsUserID (0.00s)
      logging_test.go:66: expected middleware to log the request, but the log was empty
  ```
- **Suspected file(s)**: `internal/middleware/logging.go` und `internal/middleware/logging_test.go`. Die Middleware nutzt eine eigene Paketvariable `output` und erzeugt daraus einen eigenen `log.Logger`, statt den globalen Standard-Logger zu verwenden. Die Tests fangen die Ausgabe über `log.SetOutput` am globalen Logger ab, der jedoch von der Middleware ignoriert wird. Da beide Tests identisch scheitern, liegt die gemeinsame Ursache in dieser Entkopplung. Alternativ könnten die Tests den Export `SetOutput` aus dem Middleware-Paket verwenden, der existiert, aber im Test nicht aufgerufen wird.
- **Severity**: high – `go test` endet mit Exit-Code 1, das bricht die CI/Auslieferung und verhindert die Verifikation der Logging-Anforderung (AC-11, AC-14, AC-16).