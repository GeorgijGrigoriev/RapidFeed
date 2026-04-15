# Migration: normalize feed text

## What it does

Reads every row from the `feeds` table and applies the same text normalization
that now runs at insert time:

- strips HTML tags
- decodes HTML entities (`&amp;` → `&`, `&#39;` → `'`, etc.)
- collapses whitespace and removes non-breaking spaces

Affected columns: `title`, `source`, `description`.  
Rows that are already clean are skipped (no unnecessary writes).

## When to run

**Required before upgrading to v1.0.9.**

Starting from v1.0.9 the server no longer normalizes text on every read.
Rows written by older versions may still contain raw HTML or encoded entities.
Running this migration once brings those rows in line with the new format.

## How to run

Make sure the server is stopped before running the migration.

**Via Make (recommended):**

```bash
make migrate-normalize-feeds
```

By default the migration opens `./feeds.db`.
To point it at a different database pass `DB_PATH`:

```bash
DB_PATH=/path/to/feeds.db make migrate-normalize-feeds
```

**Directly:**

```bash
go run cmd/rapidfeed/main.go -migrate-normalize-feeds
# or, if you already have a built binary:
./rapidfeed -migrate-normalize-feeds
```

## What to expect

The migration logs progress for each batch of 500 rows and prints a summary
on completion:

```
INFO [migrate] starting feed text normalization
INFO [migrate] batch done offset=0 batch_size=500 updated_so_far=312
INFO [migrate] batch done offset=500 batch_size=500 updated_so_far=589
...
INFO [migrate] feed text normalization complete total_rows=1240 rows_updated=589
```

If something goes wrong the migration exits with a non-zero code and logs the
error. It is safe to re-run: already-normalized rows are detected and skipped.
