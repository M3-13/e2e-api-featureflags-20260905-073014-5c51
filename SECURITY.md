VERDICT: CHANGES_REQUESTED

## Sicherheitsbericht

### Geprüfte Bereiche
- **Secrets:** keine hardcodierten Zugangsdaten, Tokens, Passwörter oder URLs gefunden.
- **Injection/Eingaben:** SQL-/Kommando-Injektion nicht möglich (kein SQL, keine Shell). JSON-Eingaben werden größenbegrenzt dekodiert. Verbesserung bei strikter JSON-Validierung (s. Befund 4). Log-Ausgaben werden von Steuerzeichen bereinigt; Verbesserung beim Query-Logging (s. Befund 3).
- **AuthN/AuthZ:** nicht vorhanden (s. Befund 1).
- **Dependencies:** Kein Scannerlauf anwendbar; das Projekt nutzt nur die Go-Standardbibliothek. Keine bekannten verwundbaren Pakete sichtbar.
- **Konfiguration/Transport:** HTTP ohne TLS, Bindung an alle Interfaces (`:8080`) (s. Befund 2). Keine offenen CORS-/Debug-Einstellungen.

### Bewertung
Die Acceptance Criteria sind funktional weitgehend erfüllt: Body-Limit 1 MiB, Maskierung des `user`-Parameters im Log, generische Fehlerantworten ohne Interna, keine Persistenz von Nutzer-IDs und thread-sicherer Store. Die Befunde betreffen vor allem fehlende Transport- und Zugriffssicherheit sowie Härtung der Eingabe-/Logvalidierung.

### Befunde

#### 1. Mittel – Keine Authentifizierung/Autorisierung an den CRUD- und Evaluate-Endpunkten
**Betroffen:** `main.go`, alle Handler in `internal/handlers/`

Jeder Rechner mit Netzwerkzugriff auf Port 8080 kann Flags anlegen, ändern, löschen und Evaluationsergebnisse für beliebige `user` abfragen. Insbesondere die Mutationen von Feature-Flags können das Verhalten abhängiger Anwendungen steuern.

**Fix:**
- API-Key- oder JWT-basierte Middleware vor die Handler schalten, die `Authorization` prüft.
- Falls der Dienst ausschließlich in einem bereits abgeschotteten internen Netz läuft, mindestens in der Deployment-Konfiguration dokumentieren und Port/Interface auf ein privates Interface binden (z. B. `127.0.0.1:8080` statt `:8080`).
- Der Evaluate-Endpunkt sollte zudem gegen Abfragen fremder User-IDs geschützt werden (Server-Authentifizierung des Aufrufers).

#### 2. Mittel – Transport unverschlüsselt (HTTP)
**Betroffen:** `main.go`, Zeile `http.ListenAndServe(":8080", handler)`

Der `user`-Wert wird im Query-String und Flag-Konfigurationen werden im Klartext übertragen. Die Logging-Middleware maskiert `user` im Log, das schützt jedoch nicht den Transportweg.

**Fix:**
- `http.ListenAndServeTLS` mit gültigem Zertifikat verwenden oder einen TLS-terminierenden Reverse-Proxy davorschalten.
- Wenn der Dienst nur lokal genutzt wird, Bindung auf `127.0.0.1:8080` einschränken.

#### 3. Niedrig – Query-Logging protokolliert alle Parameter außer exakt `user`
**Betroffen:** `internal/middleware/logging.go`, Funktion `sanitizedPath`

Es wird nur der Schlüssel `user` gelöscht. Andere Query-Parameter werden URL-encodiert in die Logzeile übernommen. Ein Client kann dadurch sensible Werte in die Logdatei schreiben, wenn er sie in die URL setzt. Steuerzeichen werden entfernt, aber URL-Encoding (`%0A`) bleibt als Text erhalten.

**Fix:**
- Da die Handler außer `user` keine Query-Parameter benötigen und `user` ohnehin nicht geloggt werden soll, kann der Query-String im Logging vollständig entfernt werden: `path = r.URL.Path` und kein `RawQuery` anhängen.
- Alternativ eine explizite Allowlist zulässiger Query-Parameter führen.

#### 4. Niedrig – JSON-Decoder akzeptiert nicht-EOF nach dem ersten JSON-Wert
**Betroffen:** `internal/handlers/flags.go`, Funktionen `Create` und `Update`

`json.NewDecoder(r.Body).Decode(&req)` liest nur das erste JSON-Objekt. Zusätzliche ungültige Bytes im Body werden ignoriert; z. B. `{"key":"a","enabled":true} xyz` würde 201 statt 400 liefern.

**Fix:**
Nach der ersten Dekodierung prüfen, dass kein weiteres JSON-Element folgt:

```go
dec := json.NewDecoder(r.Body)
if err := dec.Decode(&req); err != nil { /* bestehende Fehlerbehandlung */ }
if dec.More() {
    writeError(w, http.StatusBadRequest, "invalid JSON body")
    return
}
```

#### 5. Niedrig – Keine Längenbegrenzung für `key` und `description`
**Betroffen:** `internal/handlers/flags.go` (`Create`, `Update`), `internal/model/flag.go`

Das 1-MiB-Body-Limit begrenzt die Gesamtgröße, aber ein einzelner `key` oder `description` kann nahezu 1 MiB groß werden. Sehr lange Werte landen in Map, Logs und ggf. URL-Pfaden und begünstigen Log-/Ressourcen-Missbrauch.

**Fix:**
- Maximale Länge definieren, z. B. `len(req.Key) <= 256` und `len(req.Description) <= 4096`, bei Überschreitung 400 mit `writeError`.

#### 6. Niedrig – Globale Ausgabevariable der Logging-Middleware unsynchronisiert
**Betroffen:** `internal/middleware/logging.go`, `var output io.Writer`

`SetOutput` wird derzeit nur in Tests vor dem Start genutzt. Falls jemand die Variable zur Laufzeit ändert, entsteht ein Datenrennen. In der Produktion aktuell statisch.

**Fix:**
- `SetOutput` zur Konfiguration der Middleware beim Bau verwenden (z. B. `middleware.Logging(next, logger)`), statt eine globale Variable zu mutieren.
- Oder in der Dokumentation klarstellen, dass `SetOutput` nur vor dem Start aufgerufen werden darf.

### Nicht beanstandet
- `http.MaxBytesReader` wird in `Create` und `Update` vor dem Dekodieren gesetzt und `*http.MaxBytesError` korrekt auf 413 abgebildet.
- `writeError` gibt ausschließlich die vorgesehene Meldung aus; Stacktraces oder interne Fehlerdetails gelangen nicht in den Body.
- `sanitizedPath` entfernt Steuerzeichen aus Pfad und Query.
- Der In-Memory-Store verwendet `sync.RWMutex`; parallele Zugriffe sind synchronisiert.
- `?user=` wird nicht im Store persistiert, sondern nur transient für die Hash-Berechnung verwendet.
- Keine hartcodierten Secrets, keine bekannten verwundbaren Drittpakete.

**Gesamtentscheidung:** Es liegen keine kritischen oder hochkritischen Mängel vor, die einen sofortigen Stopp rechtfertigen würden. Die mittleren Befunde (fehlende Auth/AuthZ und unverschlüsselter Transport) sollten vor einem produktiven Einsatz behoben bzw. durch das Deployment kompensiert werden.