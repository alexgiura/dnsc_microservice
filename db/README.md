# Database (PostgreSQL)

## Docker și persistență

PostgreSQL rulează în Docker (serviciul `dnsc-postgres` din `docker-compose.yml`). Datele sunt stocate într-un **volume Docker**, nu în container.

În `docker-compose.yml`, volumul are **nume fix pe host: `dnsc_dnsc_pgdata`**. Astfel, la fiecare deploy rămâne **același** director de date, indiferent de numele folderului sau al proiectului Compose. Dacă s-ar folosi un volum nou (fără nume fix + alt `COMPOSE_PROJECT_NAME`), Postgres ar porni ca la prima instalare: parola din `.env` s-ar aplica la un cluster gol, iar vechea parolă din volumul vechi nu s-ar mai folosi — pare că „se schimbă parola la deploy”, deși de fapt e **alt volum / alt DB**.

- **Ștergi containerul** (`docker rm dnsc-postgres`) → volume-ul rămâne, datele nu se pierd.
- **Pornesti din nou** (`docker compose up -d`) → un container nou folosește același volume, Postgres vede datele existente.

Scripturile din `init_scripts/` (ex. `01_create_tables.sql`) rulează **doar la prima pornire**, când volume-ul e gol. După aceea nu se mai execută.

**Parola:** `POSTGRES_PASSWORD` din env setează userul doar la **prima inițializare** a volumului. Schimbarea `.env` nu actualizează singură parola în DB; folosește `ALTER USER` dacă vrei o parolă nouă fără a șterge datele.

## Comenzi utile

```bash
# Pornire (din rădăcina proiectului)
docker compose up -d

# Oprire (volume-ul dnsc_pgdata rămâne)
docker compose down

# Backup volume (producție: nume fix dnsc_dnsc_pgdata — vezi docker-compose.yml)
docker run --rm -v dnsc_dnsc_pgdata:/data -v $(pwd):/backup alpine tar czf /backup/pgdata-backup.tar.gz -C /data .
```

## Init scripts

Fișierele `.sql` din `init_scripts/` sunt montate în `/docker-entrypoint-initdb.d/` și sunt rulate în ordine alfabetică la prima inițializare a bazei.
