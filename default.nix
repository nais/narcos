{
  version,
  pkgs,
  buildGoApplication ? pkgs.buildGoApplication,
}:
buildGoApplication {
  pname = "narcos";
  version = version;
  pwd = ./.;
  src = ./.;
  modules = ./gomod2nix.toml;
  postInstall = ''
    mv $out/bin/narcos $out/bin/narc
  '';
}
