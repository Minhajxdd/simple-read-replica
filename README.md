# Simple Read Replica

Master Node Setup

This guide sets up the **read replica** for a simple streaming replication for psql db using Docker Compose.

Before starting, run the Docker Compose to create two Postgres containers:

```bash
docker compose up -d
```

Then, run the Golang HTTP server which contains a simple migration configuration that creates the `school` database:

```bash
cd simple-read-replica
go run main.go
```

Verify that the database exists:

```bash
docker exec -it postgres-master /usr/lib/postgresql/16/bin/psql -U root -d school -c "\l"
```

---

## 2 — Switch to the `postgres` user inside the container

```bash
docker exec -it --user postgres postgres-master bash
```

---

## 3 — Install nano (if needed) and edit `postgresql.conf`

```bash
apt update && apt install -y nano
nano /var/lib/postgresql/data/postgresql.conf
```

Set the following:

```conf
listen_addresses = '*'
```

Explanation:

* `listen_addresses = '*'` allows other containers on the Docker network to connect to the master.

---

## 4 — Edit `pg_hba.conf` to allow replication

```bash
nano /var/lib/postgresql/data/pg_hba.conf
```

Add:

```conf
host    replication     repuser     172.18.0.0/16     trust
```

This allows containers on the Docker network to connect as the replication user.

---

## 5 — Create the replication user

```bash
/usr/lib/postgresql/16/bin/psql -d school -U root -c "CREATE ROLE repuser WITH REPLICATION;"
```

---

## 6 — Reload or restart Postgres

```bash
/usr/lib/postgresql/16/bin/pg_ctl -D /var/lib/postgresql/data reload
# or restart if necessary
/usr/lib/postgresql/16/bin/pg_ctl -D /var/lib/postgresql/data restart
```

---

## 7 — Quick verification

Check server listening addresses:

```bash
/usr/lib/postgresql/16/bin/psql -U root -d school -c "SHOW listen_addresses;"
```

Check replication user exists:

```bash
/usr/lib/postgresql/16/bin/psql -U root -d school -c "SELECT rolname, rolreplication FROM pg_roles WHERE rolname='repuser';"
```

**Notes:**

* Use full Postgres binary paths because the Docker container’s `$PATH` may not include them.
* This master node is now ready for a simple read replica to connect using streaming replication.
* Docker `/16` subnet allows any replica container to connect without specifying individual IPs.
