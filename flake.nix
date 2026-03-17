{
  description = "wen - a natural language date CLI tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        wenVersion = "1.1.0";

        hashes = {
          x86_64-linux = "sha256-1hcriZ5Zl05iuHqQKapKY/00L19N681K3Cg0e1ZC1rI=";
          aarch64-linux = "sha256-ud0HOwded71WZ6pt5kQmD+0X+VARwYWTA99TV2VcDws=";
          x86_64-darwin = "sha256-qxp8tBMrqLnindYJpjqjVNL2cux6a6q8rIiZgR7X5C4=";
          aarch64-darwin = "sha256-4C9ASmRm6tNomPUAp1lJqsplmbRrtPlw0OlI+4EEFAU=";
        };

        archMap = {
          x86_64-linux = "linux_amd64";
          aarch64-linux = "linux_arm64";
          x86_64-darwin = "darwin_amd64";
          aarch64-darwin = "darwin_arm64";
        };

        wen-bin = pkgs.stdenv.mkDerivation {
          pname = "wen";
          version = wenVersion;

          src = pkgs.fetchurl {
            url = "https://github.com/zachthieme/wen/releases/download/v${wenVersion}/wen_${archMap.${system}}.tar.gz";
            sha256 = hashes.${system};
          };

          sourceRoot = ".";

          installPhase = ''
            mkdir -p $out/bin
            cp wen $out/bin/wen
            chmod +x $out/bin/wen
          '';

          meta = with pkgs.lib; {
            description = "A natural language date CLI tool";
            homepage = "https://github.com/zachthieme/wen";
            mainProgram = "wen";
          };
        };

        wen-src = pkgs.buildGoModule {
          pname = "wen";
          version = wenVersion;

          src = ./.;

          vendorHash = null;

          ldflags = [ "-s" "-w" ];

          meta = with pkgs.lib; {
            description = "A natural language date CLI tool";
            homepage = "https://github.com/zachthieme/wen";
            mainProgram = "wen";
          };
        };
      in
      {
        packages = {
          inherit wen-bin wen-src;
          default = wen-bin;
        };
      }
    );
}
