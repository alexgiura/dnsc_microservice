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

