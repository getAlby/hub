{
  description = "Alby Hub - Your Own Center for Internet Money";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.callPackage /mnt/data1/time2/time/2023/07/06/nixpkgs/pkgs/by-name/al/albyhub/package.nix { };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            nodejs
            yarn
            wails
          ];
        };

        apps.default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/albyhub";
        };
      }
    );
}
