#!/usr/bin/env bash

# Stop the script if any command fails
set -e

HOST="vps-hostup-01" # This assumes that ssh is configured with this host pointing to malte@tomaternas.se with a valid ssh key
REMOTE_DIR="/var/lib/tomaternas/data/"
LOCAL_DIR="./data/"

echo "Downloading data from VPS and deleting extranous local files..."
rsync -avz --delete  "$HOST:$REMOTE_DIR" "$LOCAL_DIR"

echo "Done"
