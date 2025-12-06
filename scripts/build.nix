{ writeShellApplication, go }:

writeShellApplication {
  name = "build";
  runtimeInputs = [ go ];
  text = ''
    export GOFLAGS="-buildvcs=false"
    go build -o bin/dev .
  '';
}
