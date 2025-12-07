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
          version = self.shortRev or self.dirtyShortRev or "snapshot";
          scripts = lib.packagesFromDirectoryRecursive {
            inherit callPackage;
            directory = ./nix/scripts;
          };
          dev-cli = callPackage ./nix/package.nix {
            inherit version;
          };
          shell = callPackage ./nix/shell.nix { };
        in
        {
          packages = scripts // {
            default = dev-cli;
          };
          devShells = {
            default = shell;
          };
        };
    };
}
