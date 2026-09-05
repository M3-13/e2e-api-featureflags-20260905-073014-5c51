# Feature-Flag-Service

Ein Feature-Flag-Service als REST-API in Go, ausschließlich mit `net/http` aus
der Standardbibliothek. Flags werden in einem thread-sicheren In-Memory-Store
gehalten und über CRUD-Endpunkte verwaltet; ein Evaluate-Endpunkt liefert pro
Nutzer eine deterministische Rollout-Entscheidung. Dazu Eingabevalidierung mit
sauberen Statuscodes und JSON-Fehlerobjekten sowie Zugriffs-Logging als
Middleware.

## Tech-Stack

- **Sprache:** Go (1.22+)
- **Framework:** `net/http` (Standardbibliothek)
- **Speicher:** In-Memory mit `sync.RWMutex`
- **Tests:** `go test` + `net/http/httptest`

## Installation

```sh
git clone <repository-url>
cd <repository>
```

Keine externen Abhängigkeiten — es wird nur die Go-Standardbibliothek benötigt.

## Ausführen

```sh
go run .
```

Der Server lauscht auf `http://localhost:8080`.

## Endpunkte

| Methode | Pfad                         | Antwort                                             |
|---------|------------------------------|-----------------------------------------------------|
| POST    | `/flags`                     | 201 Flag \| 400/409/413 `{"error":string}`          |
| GET     | `/flags`                     | 200 `Flag[]`                                        |
| GET     | `/flags/{key}`               | 200 Flag \| 404 `{"error":string}`                  |
| PUT     | `/flags/{key}`               | 200 Flag \| 400/404/413 `{"error":string}`          |
| DELETE  | `/flags/{key}`               | 204 \| 404 `{"error":string}`                       |
| GET     | `/flags/{key}/evaluate?user={id}` | 200 `{key,enabled,result}` \| 400/404          |
| GET     | `/healthz`                   | 200 `{"status":"ok"}`                               |

### Flag-Objekt

```json
{
  "key": "my-feature",
  "enabled": true,
  "description": "optional",
  "rollout_percent": 50
}
```

Fehlerantworten verwenden genau ein Feld: `{"error":"<Meldung>"}`.

## Tests

```sh
go test ./...
```

## Features

- CRUD-Verwaltung von Feature-Flags über eine REST-API
- Deterministische Rollout-Entscheidung pro Nutzer (Evaluate-Endpunkt)
- Thread-sicherer In-Memory-Store
- Zugriffs-Logging als Middleware
- Health-Check unter `/healthz`
