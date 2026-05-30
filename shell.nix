{
  pkgs ? import <nixpkgs> { },
}:

let
  deploy-script = pkgs.writeScriptBin "deploy" ''
    #!/usr/bin/env bash
    # Stop the script if any command fails
    set -e

    DEST_DIR=""
    BINARY_NAME="server"
    SERVICE_NAME="recept.service"

    echo "Building static Go binary for Linux..."
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $BINARY_NAME ./cmd/server

    echo "Uploading binary to VPS..."
    rsync -avz ./$BINARY_NAME recept-vps:~/recept/server

    echo "Removing binary from local machine..."
    rm $BINARY_NAME

    echo "Restarting systemd service on VPS..."
    ssh recept-vps "sudo systemctl restart $SERVICE_NAME"

    echo "Deployment complete!"
  '';
in
pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    deploy-script
  ];
  shellHook = ''
    echo "Welcome to the development shell!"
    echo "Available commands: go, deploy"
  '';
}
