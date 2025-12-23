{
  description = "snitch - a friendlier ss/netstat for humans";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      eachSystem = nixpkgs.lib.genAttrs systems;

      # go 1.25 binary derivation (required until nixpkgs ships it)
      mkGo125 = pkgs:
        let
          version = "1.25.0";
          platform = {
            "x86_64-linux" = { suffix = "linux-amd64"; hash = "sha256-KFKvDLIKExObNEiZLmm4aOUO0Pih5ZQO4d6eGaEjthM="; GOOS = "linux"; GOARCH = "amd64"; };
            "aarch64-linux" = { suffix = "linux-arm64"; hash = "sha256-Bd511plKJ4NpmBXuVTvVqTJ9i3mZHeNuOLZoYngvVK4="; GOOS = "linux"; GOARCH = "arm64"; };
            "x86_64-darwin" = { suffix = "darwin-amd64"; hash = "sha256-W9YOgjA3BiwjB8cegRGAmGURZxTW9rQQWXz1B139gO8="; GOOS = "darwin"; GOARCH = "amd64"; };
            "aarch64-darwin" = { suffix = "darwin-arm64"; hash = "sha256-VEkyhEFW2Bcveij3fyrJwVojBGaYtiQ/YzsKCwDAdJw="; GOOS = "darwin"; GOARCH = "arm64"; };
          }.${pkgs.stdenv.hostPlatform.system} or (throw "unsupported system: ${pkgs.stdenv.hostPlatform.system}");
        in
        pkgs.stdenv.mkDerivation {
          pname = "go";
          inherit version;
          src = pkgs.fetchurl {
            url = "https://go.dev/dl/go${version}.${platform.suffix}.tar.gz";
            inherit (platform) hash;
          };
          dontBuild = true;
          dontPatchELF = true;
          dontStrip = true;
          installPhase = ''
            runHook preInstall
            mkdir -p $out/{bin,share/go}
            tar -xzf $src --strip-components=1 -C $out/share/go
            ln -s $out/share/go/bin/go $out/bin/go
            ln -s $out/share/go/bin/gofmt $out/bin/gofmt
            runHook postInstall
          '';
          passthru = {
            inherit (platform) GOOS GOARCH;
          };
        };

      pkgsFor = system: import nixpkgs { inherit system; };

      mkSnitch = pkgs:
        let
          rev = self.shortRev or self.dirtyShortRev or "unknown";
          version = "nix-${rev}";
          isDarwin = pkgs.stdenv.isDarwin;
          go = mkGo125 pkgs;
          buildGoModule = pkgs.buildGoModule.override { inherit go; };
        in
        buildGoModule {
          pname = "snitch";
          inherit version;
          src = self;
          vendorHash = "sha256-fX3wOqeOgjH7AuWGxPQxJ+wbhp240CW8tiF4rVUUDzk=";
          # darwin requires cgo for libproc, linux uses pure go with /proc
          env.CGO_ENABLED = if isDarwin then "1" else "0";
          env.GOTOOLCHAIN = "local";
          # darwin: use macOS 15 SDK for SecTrustCopyCertificateChain (Go 1.25 crypto/x509)
          buildInputs = pkgs.lib.optionals isDarwin [ pkgs.apple-sdk_15 ];
          ldflags = [
            "-s"
            "-w"
            "-X snitch/cmd.Version=${version}"
            "-X snitch/cmd.Commit=${rev}"
            "-X snitch/cmd.Date=${self.lastModifiedDate or "unknown"}"
          ];
          meta = {
            description = "a friendlier ss/netstat for humans";
            homepage = "https://github.com/karol-broda/snitch";
            license = pkgs.lib.licenses.mit;
            platforms = pkgs.lib.platforms.linux ++ pkgs.lib.platforms.darwin;
            mainProgram = "snitch";
          };
        };
    in
    {
      packages = eachSystem (system:
        let pkgs = pkgsFor system; in
        {
          default = mkSnitch pkgs;
          snitch = mkSnitch pkgs;
        }
      );

      devShells = eachSystem (system:
        let
          pkgs = pkgsFor system;
          go = mkGo125 pkgs;
        in
        {
          default = pkgs.mkShell {
            packages = [ go pkgs.git pkgs.vhs ];
            env.GOTOOLCHAIN = "local";
            shellHook = ''
              echo "go toolchain: $(go version)"
            '';
          };
        }
      );

      overlays.default = final: _prev: {
        snitch = mkSnitch final;
      };
    };
}
