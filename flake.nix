{
  description = "Google Docs MCP Server";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/23d72dabcb3b12469f57b37170fcbc1789bd7457";
    nixpkgs-master.url = "github:NixOS/nixpkgs/b28c4999ed71543e71552ccfd0d7e68c581ba7e9";
    utils.url = "https://flakehub.com/f/numtide/flake-utils/0.1.102";
    batman.url = "github:amarbel-llc/batman";
  };

  outputs =
    {
      self,
      nixpkgs,
      nixpkgs-master,
      utils,
      batman,
    }:
    utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          packages =
            (with pkgs; [
              just
              jq
              nodejs_latest
            ])
            ++ [
              batman.packages.${system}.bats
              batman.packages.${system}.bats-libs
            ];
        };
      }
    );
}
