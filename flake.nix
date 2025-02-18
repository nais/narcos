{
  description = "narcos cli - for nais admins";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs, }:
    let
      version = builtins.substring 0 8
        (self.lastModifiedDate or self.lastModified or "19700101");
      # goOverlay = final: prev: {
      #   go = prev.go.overrideAttrs (old: {
      #     version = "1.23.2";
      #     src = prev.fetchurl {
      #       url = "https://go.dev/dl/go1.23.2.src.tar.gz";
      #       hash = "sha256-NpMBYqk99BfZC9IsbhTa/0cFuqwrAkGO3aZxzfqc0H8=";
      #     };
      #   });
      # };
      withSystem = nixpkgs.lib.genAttrs [
        "x86_64-linux"
        "x86_64-darwin"
        "aarch64-linux"
        "aarch64-darwin"
      ];
      withPkgs = callback:
        withSystem (system:
          callback (import nixpkgs {
            inherit system;
            # overlays = [goOverlay];
          }));
    in {
      packages = withPkgs (pkgs: rec {
        narc = pkgs.buildGoModule {
          pname = "narc";
          inherit version;
          src = ./.;
          vendorHash = "sha256-4yAMLrPePI9FxvybN0AUNXn5mFTmkiooqVLUFXp3H8c=";
        };
        default = narc;
      });

      devShells = withPkgs (pkgs: {
        default = pkgs.mkShell { buildInputs = with pkgs; [ go ]; };
      });

      formatter = withPkgs (pkgs: pkgs.nixfmt-rfc-style);
    };
}
