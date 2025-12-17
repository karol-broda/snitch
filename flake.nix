{
  description = "snitch - a friendlier ss/netstat for humans";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    systems.url = "github:nix-systems/default";
  };

  outputs = { self, nixpkgs, systems }:
    let
      supportedSystems = import systems;
      forAllSystems = f: nixpkgs.lib.genAttrs supportedSystems (system: f system);

      # go 1.25 overlay (required until nixpkgs has it)
      goOverlay = final: prev:
        let
          version = "1.25.0";
          platformInfo = {
            "x86_64-linux" = { suffix = "linux-amd64"; sri = "sha256-KFKvDLIKExObNEiZLmm4aOUO0Pih5ZQO4d6eGaEjthM="; };
            "aarch64-linux" = { suffix = "linux-arm64"; sri = "sha256-Bd511plKJ4NpmBXuVTvVqTJ9i3mZHeNuOLZoYngvVK4="; };
            "x86_64-darwin" = { suffix = "darwin-amd64"; sri = "sha256-W9YOgjA3BiwjB8cegRGAmGURZxTW9rQQWXz1B139gO8="; };
            "aarch64-darwin" = { suffix = "darwin-arm64"; sri = "sha256-VEkyhEFW2Bcveij3fyrJwVojBGaYtiQ/YzsKCwDAdJw="; };
          };
          hostSystem = prev.stdenv.hostPlatform.system;
          chosen = platformInfo.${hostSystem} or (throw "unsupported system: ${hostSystem}");
        in
        {
          go_1_25 = prev.stdenvNoCC.mkDerivation {
            pname = "go";
            inherit version;
            src = prev.fetchurl {
              url = "https://go.dev/dl/go${version}.${chosen.suffix}.tar.gz";
              hash = chosen.sri;
            };
            dontBuild = true;
            installPhase = ''
              runHook preInstall
              mkdir -p "$out"/{bin,share}
              tar -C "$TMPDIR" -xzf "$src"
              cp -a "$TMPDIR/go" "$out/share/go"
              ln -s "$out/share/go/bin/go" "$out/bin/go"
              ln -s "$out/share/go/bin/gofmt" "$out/bin/gofmt"
              runHook postInstall
            '';
            dontPatchELF = true;
            dontStrip = true;
          };
        };
    in
    {
      overlays.default = final: prev: {
        snitch = final.callPackage ./nix/package.nix { };
      };

      packages = forAllSystems (system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [ goOverlay ];
          };
        in
        {
          default = pkgs.buildGoModule {
            pname = "snitch";
            version = self.shortRev or self.dirtyShortRev or "dev";
            src = self;
            vendorHash = "sha256-BNNbA72puV0QSLkAlgn/buJJt7mIlVkbTEBhTXOg8pY=";
            env.CGO_ENABLED = 0;
            ldflags = [
              "-s" "-w"
              "-X snitch/cmd.version=${self.shortRev or "dev"}"
              "-X snitch/cmd.commit=${self.shortRev or "unknown"}"
            ];
            meta = with pkgs.lib; {
              description = "a friendlier ss/netstat for humans";
              homepage = "https://github.com/karol-broda/snitch";
              license = licenses.mit;
              platforms = platforms.linux;
              mainProgram = "snitch";
            };
          };
        }
      );

      devShells = forAllSystems (system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [ goOverlay ];
          };
        in
        {
          default = pkgs.mkShell {
            packages = [ pkgs.go_1_25 pkgs.git pkgs.vhs ];
            GOTOOLCHAIN = "local";
            shellHook = ''
              echo "go toolchain: $(go version)"
            '';
          };
        }
      );
    };
}
