#!/bin/sh

# ===============================
# ENTRYPOINT SCRIPT (entrypoint.sh)
# Purpose: To automate DB migration before starting the app
# This script is run as the container's entrypoint
# ===============================

# Exit immediately if a command fails (non-zero exit code)
# Ensures container doesn’t start if migration or app init fails
set -e

# ----------------------------------------
#Run database migration
# ----------------------------------------

echo "Running DB migration..."


# This uses the `migrate` binary (included in the image)
# -path: where migration SQL files are located inside the container
# -database: the DB connection string passed as an environment variable
# -verbose: to get detailed logs during migration
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up

# ----------------------------------------
# STEP 2: Start the main application
# ----------------------------------------

echo "Starting the application..."

# `exec "$@"` replaces the current shell with the provided command
#in Dockerfile or docker-compose, the CMD will be something like:
#   CMD ["/app/main"]
# So this line will effectively run: /app/main
exec "$@"
