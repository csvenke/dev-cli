{
  writeShellApplication,
  go,
  goreleaser,
}:

writeShellApplication {
  name = "build";
  runtimeInputs = [
    go
    goreleaser
  ];
  text = ''
    export GOFLAGS="-buildvcs=false"
    goreleaser release --snapshot --clean
  '';
}
