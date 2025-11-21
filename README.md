# Base48 Member Portal

Member portál pro hackerspace Base48 s Keycloak SSO autentizací.

**Status:** 🚧 Active Development - Fáze 3 (Admin features) dokončena

## Features

- ✅ Keycloak OIDC SSO autentizace
- ✅ Správa členských profilů s přehledem plateb a bilance
- ✅ Evidence plateb a poplatků
- ✅ Flexibilní úrovně členství
- ✅ Admin rozhraní pro správu uživatelů a rolí (filtering, sorting)
- ✅ FIO Bank integrace - automatická synchronizace plateb
- ✅ Finanční přehled - správa nespárovaných příchozích plateb
- ✅ Keycloak service account integrace pro automatizaci
- ✅ Username synchronizace z Keycloak
- ✅ Type-safe SQL (sqlc)
- ✅ Pure Go SQLite driver (bez CGO)
- 🔜 Keycloak-less mode je plánován

## Quick Start

### Prerequisites

- Go 1.21+ (testováno na 1.24.0)
- Keycloak server s nakonfigurovaným realm a clientem
- (SQLite není potřeba - používá se pure Go driver)

### Setup

1. **Clone a příprava**
```bash
git clone <repo>
cd base48-portal
cp .env.example .env
```

2. **Edituj `.env`** - viz `.env.example` pro všechny potřebné proměnné

3. **Inicializuj databázi**
```bash
mkdir -p data
# Windows (MSYS/Git Bash):
sqlite3 data/portal.db < migrations/001_initial_schema.sql
# Nebo použij DB browser nebo jiný SQL client
```

4. **Nainstaluj dependencies a vygeneruj SQL kód**
```bash
go mod tidy
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
sqlc generate
```

5. **Build a spusť server**
```bash
go build -o portal.exe cmd/server/main.go
./portal.exe
```

Server běží na `http://localhost:4848` (nebo PORT z .env)

### První přihlášení

Při prvním přihlášení existujícího uživatele přes Keycloak:
1. Systém najde uživatele podle emailu
2. Automaticky naváže `keycloak_id` z OIDC tokenu
3. Synchronizuje username z Keycloak `preferred_username`
4. Další přihlášení už probíhá přímo přes Keycloak ID

## Project Structure

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

## Development

### Regenerate SQL code
```bash
sqlc generate
```

### Run with live reload
```bash
go install github.com/air-verse/air@latest
air
```

### Build for production
```bash
go build -o portal cmd/server/main.go
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

## Admin Features

Po přihlášení jako admin (role `memberportal_admin`):

**Správa uživatelů** (`/admin/users`):
- Zobrazení všech uživatelů s Keycloak statusem a rolemi
- Filtering: state, Keycloak status, balance, search
- Sorting: ID, balance (ascending/descending)
- Inline správa rolí (assign/remove)

**Finanční přehled** (`/admin/payments/unmatched`):
- Přehled nespárovaných příchozích plateb z FIO
- Kategorizace: prázdný VS, neznámý VS, sync chyby
- Collapsible sekce pro lepší přehlednost
- Statistiky a celkové částky

**API endpointy**:
- `GET /api/admin/users` - Seznam uživatelů
- `POST /api/admin/roles/assign` - Přiřadit roli
- `POST /api/admin/roles/remove` - Odebrat roli

## Automated Tasks (Cron)

Service account umožňuje automatizované úlohy bez přihlášeného uživatele:

```bash
# Synchronizace FIO plateb (doporučeno spouštět denně)
./sync_fio_payments.exe

# Aktualizace dluhového statusu na základě bilance
go run cmd/cron/update_debt_status.go
```

Test skripty:
```bash
# Test FIO API připojení
go run cmd/test/test_fio_api.go

# Zobraz všechny uživatele v Keycloak
go run cmd/test/list_users.go

# Test přiřazení/odebrání role
TEST_USER_ID=<keycloak-user-id> go run cmd/test/test_role_assign.go
```

---

Více informací viz `SPEC.md` pro detaily o architektuře a principech.
