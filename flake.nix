{
  description = "impex";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/23.11";
    gitignore = {
      url = "github:hercules-ci/gitignore.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    xc = {
      url = "github:joerdav/xc";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, gitignore, xc }:
    let
      allSystems = [
        "x86_64-linux" # 64-bit Intel/AMD Linux
        "aarch64-linux" # 64-bit ARM Linux
        "x86_64-darwin" # 64-bit Intel macOS
        "aarch64-darwin" # 64-bit ARM macOS
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {
        inherit system;
        pkgs = import nixpkgs { inherit system; };
      });
    in
    {
      packages = forAllSystems ({ pkgs, ... }: rec {
        default = impex;

        impex = pkgs.buildGo121Module {
          name = "impex";
          src = gitignore.lib.gitignoreSource ./.;
          CGO_ENABLED = 0;
          vendorHash = null;
          flags = [
            "-trimpath"
          ];
          ldflags = [
            "-s"
            "-w"
            "-extldflags -static"
          ];
        };
      });

      # `nix develop` provides a shell containing development tools.
      devShell = forAllSystems ({ system, pkgs }:
        pkgs.mkShell {
          buildInputs = with pkgs; [
            (golangci-lint.override { buildGoModule = buildGo121Module; })
            go_1_21
            xc.packages.${system}.xc
          ];
        });

      # This flake outputs an overlay that can be used to add impex to nixpkgs.
      # Example usage:
      #
      # nixpkgs.overlays = [
      #   inputs.impex.overlays.default
      # ];
      overlays.default = final: prev: {
        impex = self.packages.${final.stdenv.system}.impex;
      };
    };
}

