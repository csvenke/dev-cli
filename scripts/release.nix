{
  writeShellApplication,
  goreleaser,
}:

writeShellApplication {
  name = "release";
  runtimeInputs = [
    goreleaser
  ];
  text = ''
    goreleaser release --clean
  '';
}
