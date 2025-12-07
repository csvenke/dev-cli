{ writeShellApplication }:

writeShellApplication {
  name = "clean";
  text = ''
    rm -rf dist/
  '';
}
