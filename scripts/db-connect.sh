#!/bin/zsh

# Define the name of the database container
DB_CONTAINER_NAME="postgres-db"

# Define variables for database parameters
DB_USER="ticketuser"
DB_PASSWORD="yourpassword"
DB_NAME="ticketdb"

# Access the database within the container
docker exec -it $DB_CONTAINER_NAME psql -U $DB_USER -d $DB_NAME