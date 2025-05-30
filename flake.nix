{
  description = "Quickly build derivations in CI";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
  };

  outputs =
    { self, nixpkgs }:
    let
      lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";
      version = builtins.substring 0 8 lastModifiedDate;
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {
      overlays.default = final: prev: {
        nix-quick-build = self.outputs.packages.${final.system}.nix-quick-build;
      };

      packages = forAllSystems (system: rec {
        nix-quick-build = nixpkgsFor.${system}.callPackage (
          {
            lib,
            buildGoModule,
            makeWrapper,
            nix-eval-jobs,
            attic-client,
          }:
          buildGoModule {
            pname = "nix-quick-build";
            inherit version;
            src = ./.;
            vendorHash = null;

            nativeBuildInputs = [
              makeWrapper
            ];

            buildInputs = [
              attic-client
              nix-eval-jobs
            ];

            postFixup = ''
              wrapProgram $out/bin/nix-quick-build \
                --prefix PATH : ${
                  lib.makeBinPath [
                    nix-eval-jobs
                    attic-client
                  ]
                }
            '';
          }
        ) { };
        default = nix-quick-build;
      });
    };
}
