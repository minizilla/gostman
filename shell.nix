{ pkgs }:

pkgs.mkShell {
  name = "gostman";
  shellHook = ''
    git config pull.rebase true
    ${pkgs.neo-cowsay}/bin/cowsay -f sage "Gostman - Postman nuances in Go Test"
  '';
  buildInputs = with pkgs; [
    editorconfig-checker
    go
  ];
}
