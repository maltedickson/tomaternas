{
  pkgs ? import <nixpkgs> { },
}:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    gcc
    sqlite
  ];

  CGO_ENABLED = "1";
}
