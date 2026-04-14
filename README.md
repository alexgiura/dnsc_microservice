## dnsc microservice (local)

### Rulezi proiectul în local
1. Treci pe branch-ul `dev`.
2. În directorul rădăcină al proiectului rulează:

```bash
docker compose -f docker-compose.local.yml up -d --build
```

### Ce rulează
- `dnsc-postgres`: PostgreSQL (date persistente în volum Docker)
- `dnsc-backend`: Go API pe portul `8080` (sau `SERVER_PORT` din `.env`)
- `dnsc-frontend`: frontend-ul servit pe `3000`

### Autentificare
După login, sesiunea e ținută într-un **cookie** (nu în localStorage). Fără login valid, apelurile către API nu merg (except o listă publică de domenii).

- **Cont de test** (la prima inițializare a DB): `admin` / `admin` — schimbă parola în producție.
- **Endpoint-uri**: login `POST /auth/login`, ieșire `POST /auth/logout`, utilizator curent `GET /auth/me`. Tabele în DB: `core.app_user`, `core.user_session` (vezi `db/init_scripts/01_create_tables.sql`).

**Setări utile în `.env` (backend):** origini permise pentru frontend (`CORS_ALLOWED_ORIGINS`, lista cu virgulă), durata sesiunii (`SESSION_TTL_DAYS`), cookie (`SESSION_COOKIE_NAME`, `SESSION_COOKIE_SECURE` pe HTTPS).

**În dev** (`npm run dev` în `vue-app`): Vite trimite `/api` și `/auth` la backend; nu e nevoie să setezi `VITE_API_BASE_URL` dacă folosești rute relative.

### DataGrip: conectare la baza de date
Pentru conexiune folosește credentialele din `.env` (ex: `POSTGRES_DB_USER`, `POSTGRES_DB_PASSWORD`, `POSTGRES_DB_NAME`).

În DataGrip setează:
- `Host`: IP-ul serverului (de exemplu `178.62.245.152`) sau `localhost` daca conectezi din aceeasi masina
- `Port`: `5432` (sau valoarea din `POSTGRES_DB_PORT` / maparea din `docker-compose`)
- `User`: `dnsc_user` (din `POSTGRES_DB_USER`)
- `Password`: `dnsc_password` (din `POSTGRES_DB_PASSWORD`)
- `Database`: `dnsc_db` (din `POSTGRES_DB_NAME`)
- `SSL`: `disable` (din `POSTGRES_DB_SSLMODE=disable`)

Notă: dacă nu te poti conecta din afara serverului, verifica firewall-ul / security group pentru portul `5432` (TCP).

