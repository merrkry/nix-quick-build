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
      nix-quick-build-package =
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
        };
    in
    {
      overlays.default = final: prev: {
        nix-quick-build = final.callPackage nix-quick-build-package { };
      };

      packages = forAllSystems (
        system:
        let
          nix-quick-build = nixpkgsFor.${system}.callPackage nix-quick-build-package { };
        in
        {
          inherit nix-quick-build;
          default = nix-quick-build;
        }
      );
    };
}
