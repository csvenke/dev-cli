{ version, buildGoModule }:

buildGoModule {
  pname = "dev";
  version = version;
  src = ../.;
  vendorHash = "sha256-6ZEO+r0ywajO2m+cVVRNWh080868fZR1QQHNW9DIBDI=";
  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
  ];
  meta = {
    mainProgram = "dev";
  };
}
