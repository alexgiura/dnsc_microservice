# Database (PostgreSQL)

## Docker și persistență

PostgreSQL rulează în Docker (serviciul `dnsc-postgres` din `docker-compose.yml`). Datele sunt stocate într-un **volume named** (`dnsc_pgdata`), nu în container.

- **Ștergi containerul** (`docker rm dnsc-postgres`) → volume-ul rămâne, datele nu se pierd.
- **Pornesti din nou** (`docker compose up -d`) → un container nou folosește același volume, Postgres vede datele existente.

Scripturile din `init_scripts/` (ex. `01_create_tables.sql`) rulează **doar la prima pornire**, când volume-ul e gol. După aceea nu se mai execută.

## Comenzi utile

```bash
# Pornire (din rădăcina proiectului)
docker compose up -d

# Oprire (volume-ul dnsc_pgdata rămâne)
docker compose down

# Backup volume (numele real e proiect_dnsc_pgdata, ex. dnsc_microservice_dnsc_pgdata)
docker run --rm -v dnsc_microservice_dnsc_pgdata:/data -v $(pwd):/backup alpine tar czf /backup/pgdata-backup.tar.gz -C /data .
```

## Init scripts

Fișierele `.sql` din `init_scripts/` sunt montate în `/docker-entrypoint-initdb.d/` și sunt rulate în ordine alfabetică la prima inițializare a bazei.
