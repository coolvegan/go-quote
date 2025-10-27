# Go Quote API

Eine einfache REST-API zum Abrufen von Zitaten aus einer Textdatei.

## Endpunkt

### GET /quotes

Gibt alle Zitate als JSON zurück.

**Authentifizierung:**
- Erforderlich: API-Key im Header `X-API-Key`
- Der Key wird über die Umgebungsvariable `API_KEY` gesetzt

## Beispiel-Request mit curl

```sh
# Erfolgreicher Request
curl -H "X-API-Key: <DEIN_API_KEY>" http://localhost:8080/quotes

# Fehler: Kein API-Key
curl http://localhost:8080/quotes

# Fehler: Falscher API-Key
curl -H "X-API-Key: wrongkey" http://localhost:8080/quotes
```

## Antwortformat

```json
[
  {
    "Author": "A. A. Milne",
    "Text": "If you live to be a hundred, I want to live to be a hundred minus one day so I never have to live without you."
  },
  ...
]
```

## Starten der API

```sh
set -x API_KEY <DEIN_API_KEY>
go run cmd/web/main.go
```

## Testen

Unit Tests sind für das Einlesen der Zitate und die Authentifizierung vorhanden:

```sh
go test ./internal/quotes

go test ./cmd/web
```

## Hinweise
- Die Zitate werden aus `misc/author-quote.txt` geladen.
- Nur Standardbibliothek wird verwendet.
- API-Key ist Pflicht für alle Requests.

# Zufälliges Zitat (ohne clientId)
curl -H "X-API-Key: <DEIN_API_KEY>" http://localhost:8080/random

# Zufälliges Zitat für clientId (kein doppeltes Zitat pro Tag)
curl -H "X-API-Key: <DEIN_API_KEY>" "http://localhost:8080/random?clientId=test123"

# Fehler: Kein API-Key
curl http://localhost:8080/random

# Fehler: Falscher API-Key
curl -H "X-API-Key: wrongkey" http://localhost:8080/random

# Fehler: Alle Zitate für clientId an einem Tag verbraucht
curl -H "X-API-Key: <DEIN_API_KEY>" "http://localhost:8080/random?clientId=test123"

{
  "Author": "A. A. Milne",
  "Text": "If you live to be a hundred, I want to live to be a hundred minus one day so I never have to live without you."
}