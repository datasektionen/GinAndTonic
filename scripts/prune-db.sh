#!/bin/zsh

# Define the name of the database container
DB_CONTAINER_NAME="postgres-db"

# Define variables for database parameters
DB_USER="ticketuser"
DB_NAME="ticketdb"

# Connect to the docker container and drop all tables
docker exec -it $DB_CONTAINER_NAME psql -U $DB_USER -d $DB_NAME -c "
SELECT 'DROP TABLE IF EXISTS \"' || tablename || '\" CASCADE;' 
FROM pg_tables
WHERE schemaname = 'public';
" | grep 'DROP TABLE' | docker exec -i $DB_CONTAINER_NAME psql -U $DB_USER -d $DB_NAME

echo "All tables dropped."