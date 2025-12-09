{ writeShellApplication, go }:

writeShellApplication {
  name = "benchmark";
  runtimeInputs = [ go ];
  text = ''
    go test -bench=. ./...
  '';
}
