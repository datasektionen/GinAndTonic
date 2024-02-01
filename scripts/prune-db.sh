#!/bin/zsh

# Define the name of the database container
DB_CONTAINER_NAME="postgres-db"

# Define variables for database parameters
DB_USER="ticketuser"
DB_PASSWORD="yourpassword"
DB_NAME="ticketdb"

# PostgreSQL code to be executed
PG_CODE=$'DO $$ DECLARE
    r RECORD;
BEGIN   
    FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = current_schema()) LOOP
        EXECUTE \'DROP TABLE IF EXISTS \' || quote_ident(r.tablename) || \' CASCADE\';
    END LOOP;
END $$;'

# Access the database within the container and execute the code
docker exec -it $DB_CONTAINER_NAME psql -U $DB_USER -d $DB_NAME -c "$PG_CODE"
