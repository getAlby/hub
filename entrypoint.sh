#!/bin/sh
# entrypoint.sh

# Check if DATABASE_URL is defined and not empty
if [ -n "$DATABASE_URL" ]; then
  export DATABASE_URI="$DATABASE_URL"
  echo "DATABASE_URI set from DATABASE_URL"
fi

# Execute the main application
exec "$@"
