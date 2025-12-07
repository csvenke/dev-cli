{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs =
    inputs@{
      self,
      flake-parts,
      nixpkgs,
      ...
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = nixpkgs.lib.systems.flakeExposed;
      perSystem =
        { pkgs, ... }:
        let
          inherit (pkgs) lib callPackage;
          scripts = lib.packagesFromDirectoryRecursive {
            inherit callPackage;
            directory = ./scripts;
          };
          version = self.shortRev or self.dirtyShortRev or "snapshot";
        in
        {
          packages = scripts // {
            default = pkgs.buildGoModule {
              pname = "dev";
              version = version;
              src = ./.;
              vendorHash = "sha256-6ZEO+r0ywajO2m+cVVRNWh080868fZR1QQHNW9DIBDI=";
              ldflags = [
                "-s"
                "-w"
                "-X main.version=${version}"
              ];
              meta = {
                mainProgram = "dev";
              };
            };
          };
          devShells = {
            default = pkgs.mkShell {
              packages = with pkgs; [
                go
                gopls
                golangci-lint
              ];
            };
          };
        };
    };
}
