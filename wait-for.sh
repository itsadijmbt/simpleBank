#!/bin/sh

# wait-for.sh
# Wait until a TCP host:port becomes reachable, then execute the next command

set -e

hostport="$1"
shift

host=$(echo "$hostport" | cut -d: -f1)
port=$(echo "$hostport" | cut -d: -f2)

echo "Waiting for $host:$port to be available..."

while ! nc -z "$host" "$port"; do
  sleep 1
done

echo "$host:$port is available. Continuing..."
exec "$@"
