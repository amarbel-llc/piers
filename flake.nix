{
  description = "Google Docs MCP Server";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/23d72dabcb3b12469f57b37170fcbc1789bd7457";
    nixpkgs-master.url = "github:NixOS/nixpkgs/b28c4999ed71543e71552ccfd0d7e68c581ba7e9";
    utils.url = "https://flakehub.com/f/numtide/flake-utils/0.1.102";
    go.url = "github:friedenberg/eng?dir=devenvs/go";
    shell.url = "github:friedenberg/eng?dir=devenvs/shell";
    batman.url = "github:amarbel-llc/batman";
  };

  outputs =
    {
      self,
      nixpkgs,
      nixpkgs-master,
      utils,
      go,
      shell,
      batman,
    }:
    utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [
            go.overlays.default
          ];
        };

        version = "1.0.0";

        piers = pkgs.buildGoApplication {
          pname = "piers";
          inherit version;
          src = ./.;
          modules = ./gomod2nix.toml;
          subPackages = [ "cmd/piers" ];

          meta = with pkgs.lib; {
            description = "MCP server for Google Docs, Sheets, and Drive";
            license = licenses.isc;
          };
        };
      in
      {
        packages = {
          default = piers;
          inherit piers;
        };

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

          inputsFrom = [
            go.devShells.${system}.default
            shell.devShells.${system}.default
          ];
        };

        apps.default = {
          type = "app";
          program = "${piers}/bin/piers";
        };
      }
    );
}
