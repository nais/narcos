{
  description = "narcos cli - for nais admins";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs =
    { self, ... }@inputs:
    inputs.flake-utils.lib.eachSystem
      [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ]
      (
        system:
        let
          version = builtins.substring 0 8 (self.lastModifiedDate or self.lastModified or "19700101");
          pkgs = import inputs.nixpkgs {
            localSystem = { inherit system; };
            overlays = [
              (
                final: prev:
                let
                  version = "1.24.2";
                  newerGoVersion = prev.go.overrideAttrs (old: {
                    inherit version;
                    src = prev.fetchurl {
                      url = "https://go.dev/dl/go${version}.src.tar.gz";
                      hash = "sha256-ncd/+twW2DehvzLZnGJMtN8GR87nsRnt2eexvMBfLgA=";
                    };
                  });
                  nixpkgsVersion = prev.go.version;
                  newVersionNotInNixpkgs = -1 == builtins.compareVersions nixpkgsVersion version;
                in
                {
                  go = if newVersionNotInNixpkgs then newerGoVersion else prev.go;
                  buildGoModule = prev.buildGoModule.override { go = final.go; };
                }
              )
            ];
          };
        in
        {
          packages = rec {
            narc = pkgs.buildGoModule {
              pname = "narc";
              inherit version;
              src = ./.;
              vendorHash = "sha256-8ooqhZbmzcrdJ0m6nijgNoQs6r0HZ9G87ABoHErzgNA=";
            };
            default = narc;
          };

          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              go
              gopls
              gotools
              go-tools
            ];
          };

          formatter = pkgs.nixfmt-rfc-style;
        }
      );
}
