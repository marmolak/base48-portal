# Keycloak Setup - Member Portal

Aplikace používá **DVA Keycloak klienty** pro různé účely + vyžaduje správnou konfiguraci **realm roles**.

## Prerekvizity

- Keycloak instance: `https://sso.base48.cz`
- Realm: `hackerspace`

---

## Krok 1: Vytvoř Realm Roles

**Keycloak** → **Realm: hackerspace** → **Realm roles** → **Create role**

| Role Name | Popis |
|-----------|-------|
| `memberportal_admin` | Administrátor member portálu |
| `active_member` | Aktivní (platící) člen |
| `in_debt` | Člen v dluhu |

---

## Krok 2: Web Application Client

**Účel**: Přihlášení uživatelů přes webový prohlížeč (Authorization Code flow)

### Vytvoření klienta

**Clients** → **Create client**

#### General Settings
```
Client ID:     member-portal-web
Client type:   OpenID Connect
```

#### Capability config
```
Client authentication:     ON
Authorization:             OFF
Standard flow:             ON  ✓
Direct access grants:      OFF
Service accounts roles:    OFF
```

#### Access settings
```
Root URL:                  https://portal.base48.cz
Home URL:                  https://portal.base48.cz
Valid redirect URIs:
  - http://localhost:4848/auth/callback
  - https://portal.base48.cz/auth/callback
Valid post logout redirect URIs:
  - http://localhost:4848
  - https://portal.base48.cz
Web origins:               *
```

### Env proměnné
Po uložení → záložka **Credentials** → zkopíruj secret:
```bash
KEYCLOAK_CLIENT_ID=member-portal-web
KEYCLOAK_CLIENT_SECRET=<secret z Credentials tab>
```

---

## Krok 3: Service Account Client

**Účel**: Automatizované úlohy (cron jobs), admin API operace

### Vytvoření klienta

**Clients** → **Create client**

#### General Settings
```
Client ID:     member-portal-service
Client type:   OpenID Connect
```

#### Capability config
```
Client authentication:     ON
Authorization:             OFF
Standard flow:             OFF
Direct access grants:      OFF
Service accounts roles:    ON  ⚠️ DŮLEŽITÉ!
```

#### Access settings
```
(nechat prázdné - service account nepotřebuje redirect URIs)
```

### Přiřazení práv

Po uložení → záložka **Service account roles** → **Assign role** → **Filter by clients** → `realm-management`:

| Role | Účel |
|------|------|
| `view-users` | Číst seznam uživatelů |
| `manage-users` | Přiřazovat/odebírat role |
| `view-realm` | Číst realm konfiguraci |
| `query-users` | Vyhledávat uživatele |

### Env proměnné
```bash
KEYCLOAK_SERVICE_ACCOUNT_CLIENT_ID=member-portal-service
KEYCLOAK_SERVICE_ACCOUNT_CLIENT_SECRET=<secret z Credentials tab>
```

---

## Krok 4: Konfigurace Role Mapperu (DŮLEŽITÉ!)

⚠️ **Bez tohoto kroku se realm roles nedostanou do ID tokenu a aplikace neuvidí role uživatelů!**

### Ověření mapperu

**Client scopes** → **roles** → **Mappers** → **realm roles**

Zkontroluj nastavení:

| Pole | Hodnota |
|------|---------|
| Name | realm roles |
| Mapper Type | User Realm Role |
| Token Claim Name | realm_access.roles |
| **Add to ID token** | **ON** ⚠️ |
| Add to access token | ON |
| Add to userinfo | ON |

Pokud **Add to ID token** je `OFF`, zapni a ulož.

### Pokud mapper neexistuje

**Client scopes** → **roles** → **Mappers** → **Add mapper** → **By configuration** → **User Realm Role**

```
Name:                         realm roles
Multivalued:                  ON
Token Claim Name:             realm_access.roles
Claim JSON Type:              String
Add to ID token:              ON
Add to access token:          ON
Add to userinfo:              ON
Add to token introspection:   ON
```

---

## Krok 5: Přiřazení rolí uživatelům

**Users** → vyber uživatele → **Role mapping** → **Assign role**

Pro adminy přiřaď:
- `memberportal_admin`
- `active_member`

---

## Kompletní .env konfigurace

```bash
# Keycloak
KEYCLOAK_URL=https://sso.base48.cz
KEYCLOAK_REALM=hackerspace

# Web client (user login)
KEYCLOAK_CLIENT_ID=member-portal-web
KEYCLOAK_CLIENT_SECRET=xxx

# Service account (automation)
KEYCLOAK_SERVICE_ACCOUNT_CLIENT_ID=member-portal-service
KEYCLOAK_SERVICE_ACCOUNT_CLIENT_SECRET=yyy
```

---

## Porovnání klientů

| Vlastnost | Web Client | Service Account |
|-----------|-----------|-----------------|
| **Účel** | Přihlášení uživatelů | Automatizace, admin API |
| **OAuth2 Flow** | Authorization Code | Client Credentials |
| **Potřebuje uživatele** | Ano | Ne |
| **Service Accounts** | OFF | **ON** |
| **Redirect URIs** | Ano | Ne |
| **V audit logu** | "User: email@..." | "service-account-member-portal-service" |

---

## Testování

### Test Web Client
```bash
go run cmd/server/main.go
# Otevři http://localhost:4848/auth/login
# Po přihlášení by měla být vidět administrace (pokud máš memberportal_admin roli)
```

### Test Service Account
```bash
curl -X POST "https://sso.base48.cz/realms/hackerspace/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=member-portal-service" \
  -d "client_secret=<tvůj-secret>"

# Měl bys dostat JSON s access_token
```

### Test cron jobu
```bash
go run cmd/cron/update_debt_status.go
```

---

## Troubleshooting

### "invalid_client"
- Špatný Client ID nebo Secret
- Špatný realm v `KEYCLOAK_REALM` (např. `master` místo `hackerspace`)
- Zkontroluj `.env` soubor

### "unauthorized_client"
- Service account: **Service accounts roles** není `ON`
- **Client authentication** není `ON`

### Uživatel se přihlásí ale nevidí admin sekci
1. Zkontroluj že má roli `memberportal_admin` v Keycloaku
2. Zkontroluj **realm roles** mapper má **Add to ID token: ON**
3. **Odhlásit a znovu přihlásit** (role se čtou z tokenu při loginu)

### "insufficient_scope" nebo "access_denied"
- Service account nemá přiřazené `realm-management` role
- Jdi do **Service account roles** a přidej `manage-users`, `view-users`

### Token expired
- OAuth2 knihovna automaticky refreshuje
- Zkontroluj že Keycloak je dostupný

---

## Architektura autentizace

```
┌─────────────────┐     Authorization Code      ┌─────────────┐
│  Web Browser    │ ◄──────────────────────────►│  Keycloak   │
│  (user login)   │         + ID Token          │  SSO        │
└────────┬────────┘                             └──────┬──────┘
         │                                             │
         │ Session Cookie                              │ Client Credentials
         ▼                                             ▼
┌─────────────────┐                             ┌─────────────┐
│  Member Portal  │ ◄───── Admin API calls ────►│  Service    │
│  Web Server     │         (manage roles)      │  Account    │
└─────────────────┘                             └─────────────┘
         │
         │ Cron jobs use
         │ Service Account
         ▼
┌─────────────────┐
│  Cron Jobs      │
│  (automation)   │
└─────────────────┘
```

---

## Checklist pro novou instalaci

- [ ] Realm `hackerspace` existuje
- [ ] Realm roles vytvořeny (`memberportal_admin`, `active_member`, `in_debt`)
- [ ] Web client `member-portal-web` vytvořen
- [ ] Service account `member-portal-service` vytvořen s `Service accounts roles: ON`
- [ ] Service account má `realm-management` role (`view-users`, `manage-users`)
- [ ] **Realm roles mapper** má `Add to ID token: ON`
- [ ] Admin uživatel má přiřazenou roli `memberportal_admin`
- [ ] `.env` aktualizován se správným realm a secrets
- [ ] Web login funguje a admin vidí administraci
- [ ] Service account token test funguje
