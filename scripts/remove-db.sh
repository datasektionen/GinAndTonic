#!/bin/zsh

# Define the name of the database container and volume
DB_CONTAINER_NAME="postgres-db"
DB_VOLUME_NAME="postgres-data"

# Stop the container
docker stop $DB_CONTAINER_NAME

# Get the id of the container
ID=$(docker ps -aqf "name=$DB_CONTAINER_NAME")

# Check the id
if [ -z "$ID" ]; then
    echo "No container with name $DB_CONTAINER_NAME"
else
    echo "Container with name $DB_CONTAINER_NAME has id $ID"
    # remove the container
    docker rm -f $ID
fi

# Clear cache and remove the volume
docker volume rm $DB_VOLUME_NAME

# Start a new container
docker compose up --build