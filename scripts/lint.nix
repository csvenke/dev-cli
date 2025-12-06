{
  writeShellApplication,
  go,
  golangci-lint,
}:

writeShellApplication {
  name = "lint";
  runtimeInputs = [
    go
    golangci-lint
  ];
  text = ''
    export GOFLAGS="-buildvcs=false"
    golangci-lint run ./...
  '';
}
