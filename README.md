# Base48 Go Member Portal

Členský portál brněnského hackerspace Base48.

**Stav:** 🚧 Aktivní vývoj.

## Fičurky

- ✅ Jednoduchá Go technologická základna
- ✅ Členům poskytuje informace a umožňuje spravovat profil a členství
- ✅ Automaticky čte a páruje platby z FIO Banky
- ✅ Automaticky řeší měsíční členské příspěvky
- ✅ Použivá Keycloak jako zdroj identit
- ✅ Správcům poskytuje administrativní webové rozraní pro správu uživatelů, plateb, fundraisingu, nastavení....
- 🔜 Email systém (uvítání, instrukce k platbě, upomínky apod...)
- 🔜 Režim fungující bez Keycloak IDP
- Viz github issues.

## Návod ke spuštění

### Předpoklady

- Go 1.21+ (testováno na 1.24.0)
- Keycloak server s nakonfigurovaným realm a clientem
- SQLite3 CLI (pro inicializaci DB)

### Nastavení a spuštění

```bash
# 1. Setup (závislosti + config)
make setup

# 2. Inicializuj databázi
make db-init

# 3. Edituj .env soubor
nano .env  # nebo tvůj editor

# 4. Vygeneruj SQL kód
make sqlc

# 5. Spusť server
make run         # jednorázové spuštění
make dev         # s hot reload (air)
```

Server běží na `http://localhost:4848` (nebo PORT z .env)

## Struktura projektu

```
base48-portal/
├── cmd/
│   ├── server/          # Main aplikace
│   ├── import/          # Import tool ze staré databáze
│   ├── cron/            # Plánované úlohy (sync_fio_payments, update_debt_status)
│   └── test/            # Test skripty pro Keycloak a FIO API
├── internal/
│   ├── auth/            # Keycloak OIDC + service account
│   ├── config/          # Environment konfigurace
│   ├── db/              # Database queries (sqlc)
│   ├── fio/             # FIO Bank API client
│   ├── handler/         # HTTP handlery
│   └── keycloak/        # Keycloak Admin API client
├── web/
│   ├── templates/       # HTML templates
│   └── static/          # CSS, JS, assets
├── migrations/          # SQL schema & migrations
├── docs/                # Dokumentace (Keycloak setup)
├── sqlc.yaml            # sqlc konfigurace
└── SPEC.md              # Detailní specifikace
```

## Keycloak Setup

Portál používá **dva Keycloak klienty**:
1. **Web client** - pro přihlášení uživatelů přes prohlížeč
2. **Service account client** - pro automatizaci (cron úlohy, admin operace)

### Web Application Client

1. Vytvoř nový Client v Keycloak:
   - Client ID: `member-portal`
   - Client Protocol: `openid-connect`
   - Access Type: `confidential`
   - Valid Redirect URIs: `http://localhost:4848/auth/callback`

2. Zkopíruj Client Secret z tab "Credentials"

### Service Account Client

1. Vytvoř další Client v Keycloak:
   - Client ID: `member-portal-service`
   - Client Protocol: `openid-connect`
   - Access Type: `confidential`
   - Service Accounts Enabled: `ON`

2. Zkopíruj Client Secret z tab "Credentials"

3. V tab "Service Account Roles", přiřaď:
   - **realm-management** → `view-users`, `manage-users`

### Nastavení rolí

V Keycloak vytvoř tyto **realm roles**:
- `active_member` - aktivní člen
- `in_debt` - člen s dluhem
- `memberportal_admin` - admin práva v portálu

Viz detaily v [`docs/KEYCLOAK_SETUP.md`](docs/KEYCLOAK_SETUP.md)

## Vývoj

```bash
make dev          # Run s hot reload (air)
make sqlc         # Regenerate SQL code
make build        # Build aplikace
make build-all    # Build všech binárků (server + cron)
make test         # Spusť testy
make clean        # Vymaž build artifacts
make help         # Zobraz všechny dostupné příkazy
```

## Database Schema

- **levels** - Úrovně členství (Student, Regular, Sponsor...)
- **users** - Členové hackerspace
- **payments** - Evidence plateb
- **fees** - Měsíční poplatky

Detaily viz `migrations/001_initial_schema.sql`

## Tech Stack

- **Go 1.24** - Backend
- **Chi** - HTTP router
- **go-oidc** - Keycloak OIDC autentizace
- **sqlc** - Type-safe SQL code generation
- **modernc.org/sqlite** - Pure Go SQLite driver (bez CGO)
- **Tailwind CSS** - Styling (plánováno)
- **html/template** - Server-side rendering

## Automated Tasks (Cron)

Service account umožňuje automatizované úlohy bez přihlášeného uživatele:

```bash
# Build cron jobs
make build-all

# Synchronizace FIO plateb (doporučeno spouštět denně)
./sync_fio_payments
```
---

Více informací viz `SPEC.md` pro detaily o architektuře a principech.
