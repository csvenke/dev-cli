{
  writeShellApplication,
  goreleaser,
}:

writeShellApplication {
  name = "create-release";
  runtimeInputs = [
    goreleaser
  ];
  text = ''
    goreleaser release --clean
  '';
}
