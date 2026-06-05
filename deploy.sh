#!/usr/bin/env bash

# Stop the script if any command fails
set -e

HOST="vps-hostup-01" # This assumes that ssh is configured with this host pointing to malte@tomaternas.se with a valid ssh key
LOCAL_BINARY_NAME="server"
TMP_PATH="/tmp/tomaternas-server-new"

echo "Building..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $LOCAL_BINARY_NAME ./cmd/server

echo "Uploading..."
rsync -avz ./$LOCAL_BINARY_NAME "$HOST:$TMP_PATH"


echo "Installing and restarting..."
ssh "$HOST" "
    sudo install -o tomaternas -g tomaternas -m 755 $TMP_PATH /var/lib/tomaternas/server
    sudo systemctl restart tomaternas
"

echo "Service status:"
ssh "$HOST" "systemctl is-active tomaternas"

echo "Cleaning up..."
ssh "$HOST" "rm $TMP_PATH"
rm $LOCAL_BINARY_NAME

echo "Done"
