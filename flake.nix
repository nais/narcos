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
                  version = "1.25.1";
                  newerGoVersion = prev.go_latest.overrideAttrs (old: {
                    inherit version;
                    src = prev.fetchurl {
                      url = "https://go.dev/dl/go${version}.src.tar.gz";
                      hash = "sha256-0BDBCc7pTYDv5oHqtGvepJGskGv0ZYPDLp8NuwvRpZQ=";
                    };
                  });
                  nixpkgsVersion = prev.go_latest.version;
                  newVersionNotInNixpkgs = -1 == builtins.compareVersions nixpkgsVersion version;
                in
                {
                  go_latest = if newVersionNotInNixpkgs then newerGoVersion else prev.go_latest;
                  buildGoLatestModule = prev.buildGoLatestModule.override { go = final.go_latest; };
                }
              )
            ];
          };
        in
        {
          packages = rec {
            narc = pkgs.buildGoLatestModule {
              pname = "narc";
              inherit version;
              src = ./.;
              vendorHash = "sha256-ZLmOTeEXAdmjl0Ls9PSko3c5ZCz6w6+jo1QNANB0DN4=";
            };
            default = narc;
          };

          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              go_latest
              gopls
              gotools
              go-tools
            ];
          };

          formatter = pkgs.nixfmt-rfc-style;
        }
      );
}
