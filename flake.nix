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
          x86_64-linux = "sha256-NLqP7T86BBDJly2S4vhyS+VJ4hVcLgGCUGM7koilCr8=";
          aarch64-linux = "sha256-zrWrZU7tSaZ0NMdfovqqFyH5aUhX01OcR2Oayzix1Uc=";
          x86_64-darwin = "sha256-oZidOUDsqm8PSmsbcQbjninMaTDMyS0Z4iCVc40oYqg=";
          aarch64-darwin = "sha256-lI8Jt1Ff/I2Sp/5fRCaynyZSy12NlHVmf7YgJ72Xs/g=";
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
