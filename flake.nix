{
  description = "Ren Browser - Reticulum browser for NomadNet pages";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        lib = pkgs.lib;

        version =
          let
            brand = builtins.readFile ./build/brand.yml;
            m = builtins.match ".*\nversion: \"([^\"]+)\".*" brand;
          in
          if m == null then "0.2.0" else builtins.elemAt m 0;

        gitCommit =
          if self ? shortRev then
            self.shortRev
          else if self ? dirtyShortRev then
            self.dirtyShortRev
          else
            "dev";

        commonNative = with pkgs; [
          pkg-config
          nodejs_22
          pnpm
          go-task
          git
          makeWrapper
        ];

        guiLibs = with pkgs; [
          gtk4
          webkitgtk_6_0
        ];

        # First `nix build` may fail and print the correct pnpmDeps hash — paste it here.
        pnpmDeps = pkgs.fetchPnpmDeps {
          pname = "renbrowser";
          inherit version;
          src = ./.;
          fetcherVersion = 3;
          hash = "sha256-DJOCvPpah/A7mKN1XWUTZIZ+B0/WrfTh9cf0+/quivE=";
          pnpm = pkgs.pnpm;
          pnpmWorkspaces = [ "renbrowser-frontend" ];
        };

        frontend = pkgs.stdenv.mkDerivation {
          pname = "renbrowser-frontend";
          inherit version;
          src = ./.;

          nativeBuildInputs = with pkgs; [
            nodejs_22
            pnpm
            pnpmConfigHook
          ];

          inherit pnpmDeps;
          pnpmRoot = ".";

          buildPhase = ''
            runHook preBuild
            export HOME="$TMPDIR"
            pnpm --dir frontend run build
            runHook postBuild
          '';

          installPhase = ''
            runHook preInstall
            mkdir -p "$out"
            cp -a frontend/dist "$out/"
            runHook postInstall
          '';
        };

        mkGoApp =
          {
            pname,
            tags,
            cgo ? true,
            buildInputs ? [ ],
            postInstall ? "",
            meta,
          }:
          pkgs.buildGoModule {
            inherit
              pname
              version
              meta
              ;
            src = ./.;

            # Go modules are vendored in-tree (including third_party/reticulum-go).
            vendorHash = null;
            proxyVendor = false;

            nativeBuildInputs = commonNative;
            inherit buildInputs;

            preBuild = ''
              rm -rf frontend/dist
              mkdir -p frontend
              cp -a ${frontend}/dist frontend/dist
              # nixpkgs Go may lag the go.mod patch (e.g. 1.26.4 vs 1.26.5).
              sed -i 's/^go .*/go ${pkgs.go.version}/' go.mod
              bash build/scripts/patch-wails-vendor.sh
            '';

            env = {
              CGO_ENABLED = if cgo then "1" else "0";
              GOFLAGS = "-mod=vendor";
              GOTOOLCHAIN = "local";
            };

            subPackages = [ "." ];
            inherit tags;

            ldflags = [
              "-s"
              "-w"
              "-X renbrowser/internal/buildinfo.Version=${version}"
              "-X renbrowser/internal/buildinfo.Commit=${gitCommit}"
            ];

            doCheck = false;
            inherit postInstall;
          };

        renbrowser = mkGoApp {
          pname = "renbrowser";
          tags = [ "production" ];
          cgo = true;
          buildInputs = guiLibs;
          postInstall = ''
            wrapProgram "$out/bin/renbrowser" \
              --prefix XDG_DATA_DIRS : "$GSETTINGS_SCHEMAS_PATH"

            install -Dm644 flatpak/io.quad4.renbrowser.desktop \
              "$out/share/applications/io.quad4.renbrowser.desktop"
            sed -i \
              -e 's|^Exec=.*|Exec=renbrowser|' \
              -e 's|^Icon=.*|Icon=io.quad4.renbrowser|' \
              "$out/share/applications/io.quad4.renbrowser.desktop"
            install -Dm644 flatpak/io.quad4.renbrowser.svg \
              "$out/share/icons/hicolor/scalable/apps/io.quad4.renbrowser.svg"
            install -Dm644 LICENSE "$out/share/licenses/renbrowser/LICENSE"
          '';
          meta = with lib; {
            description = "Reticulum browser for NomadNet pages";
            homepage = "https://github.com/Quad4-Software/Ren-Browser";
            license = licenses.mit;
            mainProgram = "renbrowser";
            platforms = platforms.linux;
          };
        };

        renbrowser-server = mkGoApp {
          pname = "renbrowser-server";
          tags = [
            "server"
            "production"
          ];
          cgo = false;
          buildInputs = [ ];
          postInstall = ''
            mv "$out/bin/renbrowser" "$out/bin/renbrowser-server"
            install -Dm644 LICENSE "$out/share/licenses/renbrowser-server/LICENSE"
          '';
          meta = with lib; {
            description = "Headless Ren Browser server";
            homepage = "https://github.com/Quad4-Software/Ren-Browser";
            license = licenses.mit;
            mainProgram = "renbrowser-server";
            platforms = platforms.unix;
          };
        };
      in
      {
        packages = {
          default = renbrowser;
          inherit renbrowser renbrowser-server frontend;
        };

        apps.default = {
          type = "app";
          program = lib.getExe renbrowser;
        };

        devShells.default = pkgs.mkShell {
          packages =
            commonNative
            ++ guiLibs
            ++ (with pkgs; [
              go
              gopls
              gotools
              gcc
            ]);

          shellHook = ''
            export CGO_ENABLED=1
            export GOFLAGS="-mod=vendor"
            echo "Ren Browser nix shell"
            echo "  task build"
            echo "  task package:linux:arch"
            echo "  nix build"
          '';
        };

        formatter = pkgs.nixfmt-rfc-style;
      }
    );
}
