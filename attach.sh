#!/bin/bash

# Check if a container ID is provided
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <container-id>"
    exit 1
fi

CONTAINER_ID=$1

# Attach to the container
docker exec -it $CONTAINER_ID /bin/sh
