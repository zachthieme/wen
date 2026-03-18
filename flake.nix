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

        wenVersion = "1.2.1";

        hashes = {
          x86_64-linux = "sha256-8kWWG05VVrSyQswLqOAI32zb4nW0ft6USxJvKBbxWm0=";
          aarch64-linux = "sha256-VpW/8Mw0lRuBHyDGOJG/dpZkCzgyQMKivL/hT7UF1oI=";
          x86_64-darwin = "sha256-6udeDDFjW1Agi9xmIFDgf3KhzEJbcCdv3OYKJym7BoE=";
          aarch64-darwin = "sha256-Dw/7UvFi4jVI99Sq5fNciZ5+j27LujZiUUuP1dmLEHY=";
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
