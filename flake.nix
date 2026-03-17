{
  description = "zdate - a natural language date CLI tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        zdateVersion = "0.1.0";

        hashes = {
          x86_64-linux = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
          aarch64-linux = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
          x86_64-darwin = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
          aarch64-darwin = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
        };

        archMap = {
          x86_64-linux = "linux_amd64";
          aarch64-linux = "linux_arm64";
          x86_64-darwin = "darwin_amd64";
          aarch64-darwin = "darwin_arm64";
        };

        zdate-bin = pkgs.stdenv.mkDerivation {
          pname = "zdate";
          version = zdateVersion;

          src = pkgs.fetchurl {
            url = "https://github.com/zachthieme/zdate/releases/download/v${zdateVersion}/zdate_${archMap.${system}}.tar.gz";
            sha256 = hashes.${system};
          };

          sourceRoot = ".";

          installPhase = ''
            mkdir -p $out/bin
            cp zdate $out/bin/zdate
            chmod +x $out/bin/zdate
          '';

          meta = with pkgs.lib; {
            description = "A natural language date CLI tool";
            homepage = "https://github.com/zachthieme/zdate";
            mainProgram = "zdate";
          };
        };

        zdate-src = pkgs.buildGoModule {
          pname = "zdate";
          version = zdateVersion;

          src = ./.;

          vendorHash = null;

          ldflags = [ "-s" "-w" ];

          meta = with pkgs.lib; {
            description = "A natural language date CLI tool";
            homepage = "https://github.com/zachthieme/zdate";
            mainProgram = "zdate";
          };
        };
      in
      {
        packages = {
          inherit zdate-bin zdate-src;
          default = zdate-bin;
        };
      }
    );
}
