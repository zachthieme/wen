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

        wenVersion = "1.8.0";

        hashes = {
          x86_64-linux = "sha256-1Dde3BeUJq36b6maHTIfduB6JlIc0pxfUfelmCigLe4=";
          aarch64-linux = "sha256-eSHk9vxVutYOqEff+VB9wRulgmWxWrQLPF/rPl6c6XI=";
          x86_64-darwin = "sha256-1G0y6VEeXVj7Uyyf5Dny5boJ/EDxYxBWQy05Olj9LGo=";
          aarch64-darwin = "sha256-QcFjoPoy6RMFiCYiAlHf+t66JGjU+geghoesHC8L34Y=";
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

          subPackages = [ "cmd/wen" ];

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
