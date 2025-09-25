# Simple Read Replica

This guide walks you through setting up a **Postgres master node** and a **replica (follower)** using Docker Compose with streaming replication.

---

## 1 — Bring up the containers

Start your Postgres containers with Docker Compose:

```bash
cd simple-read-replica

docker compose up -d
```

---

## 2 — Create a sample database

run the command to create a sample database 

```bash
docker exec -it postgres-master psql -U root -d school

CREATE TABLE IF NOT EXISTS records (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

\q
```

Verify that the database exists:

```bash
docker exec -it postgres-master /usr/lib/postgresql/16/bin/psql -U root -d school -c "\l"
```

---

## 3 — Configure the master node

```bash
docker exec -it postgres-master bash
```

### Install nano and edit configs

```bash
apt update && apt install -y nano

nano /var/lib/postgresql/data/postgresql.conf
```

Set:

```conf
listen_addresses = '*'
```

This makes Postgres listen on all network interfaces (needed for replication).

### Update `pg_hba.conf`

```bash
nano /var/lib/postgresql/data/pg_hba.conf
```

Add:

```conf
host    replication     repuser     172.18.0.0/16     trust
```

This lets other containers in the same Docker network connect as the replication user.

### Create the replication user

```bash
su - postgres

/usr/lib/postgresql/16/bin/psql -d school -U root -c "CREATE ROLE repuser WITH REPLICATION LOGIN;"
```

### Reload or restart Postgres

```bash
/usr/lib/postgresql/16/bin/pg_ctl -D /var/lib/postgresql/data reload
# or
/usr/lib/postgresql/16/bin/pg_ctl -D /var/lib/postgresql/data restart
```

### Quick verification

Check if Postgres is listening correctly:

```bash
/usr/lib/postgresql/16/bin/psql -U root -d school -c "SHOW listen_addresses;"
```

Check the replication user:

```bash
/usr/lib/postgresql/16/bin/psql -U root -d school -c "SELECT rolname, rolreplication FROM pg_roles WHERE rolname='repuser';"
```

At this point, the **master node is ready**.

exit from the container
```bash
exit
exit
```

---

## 4 — Set up the replica (follower)

### Stop the Postgres service in the replica container

```bash
docker exec -it --user postgres postgres-replica /usr/lib/postgresql/16/bin/pg_ctl -D /var/lib/postgresql/data stop
```

### Clean the old data directory

```bash
docker exec -it postgres-replica bash -c "rm -rf /var/lib/postgresql/data/*"
```

### Make sure the folder is empty

```bash
docker exec -it postgres-replica ls -la /var/lib/postgresql/data
```

### Run `pg_basebackup` to sync with the master

```bash
//if not working the run the remove command again and quickly run this one as well
docker exec -it postgres-replica pg_basebackup -h postgres-master -U repuser --checkpoint=fast -D /var/lib/postgresql/data/ -R --slot=some_name -C --port=5432
```

This copies data from the master into the replica and sets up replication automatically.

### Validate

Check again that files exist:

```bash
docker exec -it postgres-replica ls -la /var/lib/postgresql/data
```

Now start the replica’s Postgres (it usually starts on its own when the container restarts, but you can also run pg\_ctl manually if needed).

---

`Simply run the http sever by running the main.go file and use create and read both uses different db connection`
`or feel free to inserting and reading clusters manually as well :)`

* The Docker network (`172.18.0.0/16`) makes it easy for the master and replica to talk without fiddling with individual IPs.
* This is a simple setup — good for learning and local dev. For production, you’d want stronger auth, replication slots, monitoring, etc...
