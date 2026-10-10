#!/usr/bin/env bash
# Deletes trips (and their items) from the local database, to clear out
# trips created while testing. Basics and family members are not touched.
#
# Usage: scripts/reset-trips.sh [--keep 2,9]
#   or:  make reset-trips KEEP=2,9
#
# Safety: lists what will be deleted and asks before doing anything, takes a
# backup to backups/ first, and deletes in one transaction (all or nothing).
# Restore a backup by stopping the app and copying it over the database file.
set -euo pipefail

db="${DB_PATH:-roundtrip.db}"
keep=""

while [ $# -gt 0 ]; do
	case "$1" in
	--keep)
		keep="${2:-}"
		shift 2
		;;
	*)
		echo "unknown argument: $1" >&2
		exit 2
		;;
	esac
done

# keep is placed into SQL as-is, so only digits and commas are allowed.
if [ -n "$keep" ] && ! [[ "$keep" =~ ^[0-9]+(,[0-9]+)*$ ]]; then
	echo "--keep must be trip IDs separated by commas, e.g. --keep 2,9" >&2
	exit 2
fi

# sqlite3 creates an empty database when the file is missing, so check first.
if [ ! -f "$db" ]; then
	echo "database not found: $db" >&2
	exit 1
fi

sql() {
	sqlite3 -bail -cmd ".timeout 5000" "$db" "$@"
}

if [ -n "$keep" ]; then
	# Catch typos: a kept ID that does not exist is most likely a wrong number.
	for id in ${keep//,/ }; do
		if [ "$(sql "SELECT count(*) FROM trip WHERE id = $id;")" = "0" ]; then
			echo "no trip with ID $id, nothing deleted" >&2
			exit 1
		fi
	done
	where="WHERE id NOT IN ($keep)"
	item_where="WHERE trip_id NOT IN ($keep)"
else
	where=""
	item_where=""
fi

trips=$(sql "SELECT count(*) FROM trip $where;")
if [ "$trips" = "0" ]; then
	echo "No trips to delete."
	exit 0
fi

echo "Database: $db"
echo "Trips to delete:"
sql -separator "  " "SELECT id, departure_date, destination FROM trip $where ORDER BY id;" | sed 's/^/  /'
if [ -n "$keep" ]; then
	echo "Trips to keep:"
	sql -separator "  " "SELECT id, departure_date, destination FROM trip WHERE id IN ($keep) ORDER BY id;" | sed 's/^/  /'
fi
echo "Items to delete: $(sql "SELECT count(*) FROM item $item_where;")"

read -r -p "Delete $trips trip(s)? [y/N] " answer
if [ "$answer" != "y" ] && [ "$answer" != "Y" ]; then
	echo "Cancelled, nothing deleted."
	exit 0
fi

# .backup is safe while the app is running, unlike copying the file, which
# can miss changes still in the WAL file.
mkdir -p backups
backup="backups/$(basename "$db" .db)-$(date +%Y%m%d-%H%M%S).db"
sql ".backup '$backup'"
echo "Backup: $backup"

# Items are deleted explicitly rather than relying on ON DELETE CASCADE, which
# the sqlite3 CLI only honours with foreign_keys on. foreign_keys stays on so a
# future table that references trip makes this fail instead of leaving orphans.
sql "PRAGMA foreign_keys = ON;
BEGIN;
DELETE FROM item $item_where;
DELETE FROM trip $where;
COMMIT;"

echo "Deleted $trips trip(s). Trips left: $(sql "SELECT count(*) FROM trip;")"
