{ version, buildGoModule }:

buildGoModule {
  pname = "dev";
  version = version;
  src = ../.;
  vendorHash = "sha256-tN00iGt9fVunFvlJkURIhjqPFqPBkEq78ehR9LWcces=";
  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
  ];
  meta = {
    mainProgram = "dev";
  };
}
