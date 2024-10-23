{
  version,
  pkgs,
  buildGoApplication ? pkgs.buildGoApplication,
}:
buildGoApplication {
  pname = "narcos";
  version = version;
  subPackages = ["cmd/narc"];
  pwd = ./.;
  src = ./.;
  modules = ./gomod2nix.toml;
}
