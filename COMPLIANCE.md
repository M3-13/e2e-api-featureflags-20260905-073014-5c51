VERDICT: CHANGES_REQUESTED

# Konformitätsprüfung Feature-Flag-Service (Go-Backend)

Geprüft wurde ausschließlich der sichtbare Stand des Produkts. Reine Backend-API ohne Endnutzer-UI; daher sind **Pflichttexte, Cookie-Banner und Barrierefreiheit** nicht anwendbar.

---

## 1. DSGVO

### G-01 — Logging der Query-Strings kann künftige/weitere personenbezogene Parameter erfassen
- **Schweregrad:** mittel
- **Fundstelle:** `internal/middleware/logging.go`, Funktion `sanitizedPath`
- **Sachverhalt:** Der Parameter `user` wird korrekt aus der Log-Ausgabe entfernt. Andere Query-Parameter bleiben jedoch erhalten und werden protokolliert. Falls künftig oder durch Client-Fehler weitere Parameter wie `email`, `token` oder Kundennummern übergeben werden, landen diese ungefiltert im Log. Das widerspricht der Datenminimierung (Art. 5 Abs. 1 lit. c DSGVO) und der wörtlichen Maßgabe aus AC-16, den Query-String zu entfernen oder vollständig zu maskieren.
- **Abhilfe:** In `sanitizedPath` den kompletten Query-String entfernen:
  ```go
  func sanitizedPath(r *http.Request) string {
      return stripControl(r.URL.Path)
  }
  ```
  Falls einzelne unkritische Parameter protokolliert werden sollen, eine explizite Allowlist verwenden. Die API-Funktion wird dadurch nicht beeinträchtigt; nur das Log wird sparsamer.

### G-02 — Betreiberdokumentation zur Rechtsgrundlage und Datenverarbeitung fehlt
- **Schweregrad:** niedrig bis mittel
- **Fundstelle:** Projektwurzel, insbesondere `README.md` (Inhalt nicht einsehbar, daher als nicht nachweisbar gewertet)
- **Sachverhalt:** Der Dienst verarbeitet mit `?user=` technisch eine Nutzerkennung. Der Code speichert und protokolliert sie nicht — das ist datenschutzfreundlich umgesetzt. Eine Betreiberdokumentation, die die Rechtsgrundlage (z. B. Art. 6 Abs. 1 lit. f DSGVO für die serverseitige Rollout-Steuerung oder ein Auftragsverarbeitungsverhältnis) sowie die ausschließlich transiente Verarbeitung beschreibt, ist im sichtbaren Stand nicht vorhanden.
- **Abhilfe:** In `README.md` ein Kapitel „Datenschutz und Betrieb“ ergänzen:
  - `?user=` wird ausschließlich flüchtig zur Hash-Berechnung verwendet,
  - wird nicht gespeichert, nicht protokolliert und nicht an Dritte übergeben,
  - der Betreiber muss eine Rechtsgrundlage sicherstellen,
  - Löschkonzept ist trivial, da keine Nutzerdaten persistiert werden.

---

## 2. EU Cyber Resilience Act (CRA)

### C-01 — Administrative Endpunkte ohne Authentifizierung und Autorisierung
- **Schweregrad:** **hoch**
- **Fundstelle:** `main.go` (Registrierung von `POST /flags`, `PUT /flags/{key}`, `DELETE /flags/{key}`)
- **Sachverhalt:** Jeder mit Netzzugang kann Flags anlegen, ändern und löschen. Der Standardbetrieb lauscht auf `:8080` an allen Interfaces. Das ist kein „secure by design/default“ gemäß CRA Anhang I. Die Verfügbarkeit und Integrität des Dienstes ist ohne weitere Schutzmaßnahmen nicht kontrolliert.
- **Abhilfe:** In `internal/middleware` eine Authentifizierungs-Middleware für die Schreib-Endpunkte ergänzen, z. B.:
  - `Authorization: Bearer <token>` prüfen,
  - Token aus `os.Getenv("FLAG_ADMIN_TOKEN")` lesen,
  - bei fehlender Konfiguration für Admin-Routen fail-closed verhalten oder explizit einen unsicheren Entwicklungsmodus verlangen.
  Alternativ im `README.md` verbindlich dokumentieren, dass der Dienst ausschließlich hinter einem authentifizierenden Reverse Proxy betrieben werden darf. Die bestehenden Handler-Tests können weiterhin direkt gegen die Handler laufen, wenn die Middleware nur in `main.go` für die Produktionsroute aktiviert wird.

### C-02 — Kein TLS und keine konfigurierbaren Server-Timeouts
- **Schweregrad:** mittel
- **Fundstelle:** `main.go`, `http.ListenAndServe(":8080", handler)`
- **Sachverhalt:** Der Dienst bietet keine Transportverschlüsselung, bindet fest an Port `8080` und nutzt den Default-Server ohne Lese-/Schreib-Timeouts. Das ist für einen sicherheitsrelevanten Verwaltungsdienst nicht ausreichend.
- **Abhilfe:** In `main.go` einen expliziten `http.Server` mit `ReadTimeout`, `WriteTimeout`, `IdleTimeout` und konfigurierbarer Adresse (`ADDR`) erzeugen. Optional TLS über `TLS_CERT_FILE`/`TLS_KEY_FILE` unterstützen oder im `README.md` TLS-Terminierung durch einen vorgelagerten Proxy vorschreiben. Lese- und Schreib-Timeouts schützen zusätzlich vor langsamen Verbindungen.

### C-03 — Kein SBOM, keine Versionsangabe, kein dokumentierter Update-/Patch-Prozess
- **Schweregrad:** niedrig bis mittel
- **Fundstelle:** Projektwurzel / `go.mod`; kein sichtbares SBOM-Manifest
- **Sachverhalt:** Das Produkt besteht derzeit nur aus der Go-Standardbibliothek, daher ist das Lieferkettenrisiko gering. Ein nachweisbarer SBOM, eine Produktversion und ein dokumentierter Update-/Patchprozess fehlen jedoch im sichtbaren Stand. CRA verlangt für Produkte mit digitalen Elementen dokumentierte Sicherheitseigenschaften und eine Update-Fähigkeit.
- **Abhilfe:** 
  - In `internal/version` eine Versionskonstante pflegen und in `GET /healthz` als `version` ausgeben.
  - Im Build-Prozess einen SBOM generieren (z. B. `go version -m` oder CycloneDX).
  - Im `README.md` oder `AGENTS.md` dokumentieren, wie Updates eingespielt werden, wie Schwachstellen gemeldet werden und welche Support-/Patch-Frist gilt.

### C-04 — Log-Härtung unvollständig und fehlende Header-Härtung
- **Schweregrad:** niedrig
- **Fundstelle:** `internal/middleware/logging.go`, `internal/handlers/response.go`
- **Sachverhalt:** 
  - `stripControl` entfernt nur ASCII-Steuerzeichen unter `0x20` und `0x7f`. Unicode-Zeilenumbrüche wie U+2028/U+2029 können Log-Viewer verwirren oder Log-Zeilen aufbrechen.
  - `writeJSON` setzt keinen `X-Content-Type-Options: nosniff`-Header.
- **Abhilfe:** 
  - `stripControl` zusätzlich `r == 0x2028 || r == 0x2029` ausfiltern.
  - In `writeJSON` `w.Header().Set("X-Content-Type-Options", "nosniff")` ergänzen.

### C-05 — Keine Längenbeschränkung für Flag-Schlüssel
- **Schweregrad:** niedrig
- **Fundstelle:** `internal/handlers/flags.go`, `createFlagRequest.Key`
- **Sachverhalt:** Der Flag-Schlüssel darf bis zur Body-Grenze beliebig lang sein und wird in Map, Antwort und Pfad verwendet. Eine sehr lange Zeichenkette erhöht das Risiko für Log-/Antwortaufblähung und erschwert die Administration.
- **Abhilfe:** In `Create` nach dem JSON-Decoding prüfen, z. B. `len(req.Key) <= 128`; bei Überschreitung `writeError(w, http.StatusBadRequest, "key too long")` zurückgeben.

---

## 3. EU AI Act

- **Befund:** keine KI-Funktion implementiert.
- Der Dienst enthält keine KI-Systeme oder KI-Modelle. Es bestehen keine Pflichten aus dem AI Act.

---

## 4. Pflichttexte & UI

- **Befund:** nicht anwendbar.
- Der Dienst ist eine reine REST-API ohne Endnutzer-UI. Es gibt keine Cookies, kein Tracking und keine Verkaufs- oder Widerrufsbeziehung gegenüber Endnutzern. Rechtstexte im Sinne von Impressum, Datenschutzerklärung oder Cookie-Banner sind auf Produktebene nicht erforderlich. Die Betreiberdokumentation ist unter G-02 angesprochen.

---

## 5. Barrierefreiheit

- **Befund:** nicht anwendbar.
- Es existiert keine öffentliche Web-UI. WCAG/BITV/EAA-Verpflichtungen greifen daher nicht.

---

## Reconcile-Hinweis

Die vorgeschlagenen Maßnahmen sind mit der Produktfunktion vereinbar:

- Das Entfernen des Query-Strings betrifft nur das Logging und nicht die API-Antworten.
- Eine Authentifizierungs-Middleware kann in `main.go` nur für Produktionsrouten aktiviert werden; Handler-Tests laufen weiterhin direkt.
- TLS und Timeouts sind konfigurierbar ausgelegt, sodass lokale Entwicklung und Tests ohne Zertifikat möglich bleiben.
- Die zusätzliche Schlüssellängen-Prüfung ist rückwärtskompatibel zu den bestehenden Anforderungen.