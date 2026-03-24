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

        wenVersion = "1.5.1";

        hashes = {
          x86_64-linux = "sha256-VBBFJ7kL0D59QcxGomYOjIJZHNGtxkVTnHaOgHgh4A8=";
          aarch64-linux = "sha256-btXPbI1nxmLrG0QcTw0wU1GrxgBHkGVQ/i+NIvZmpXQ=";
          x86_64-darwin = "sha256-OKE/NYCd3A4Jr1zGQp1fZCNc8YnpHkQUQpxvcgZJMP0=";
          aarch64-darwin = "sha256-7camqkbuOYTfUodUgvZzsz1X9I1tHOOBea1uQLTaHoA=";
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
