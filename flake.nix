{
  description = "@ctfbot_ - Stacker News CTF bot";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};

      runtimeTools = with pkgs; [
        bashInteractive
        coreutils
        gnugrep
        gnused
        gawk
        findutils
        less
        file
        gzip
        diffutils
        which
        git
        # pdf tools
        pkgs."poppler-utils"
      ];

      # The command set mounted read-only into the sandbox (SANDBOX_TOOLS).
      sandboxTools = pkgs.buildEnv {
        name = "ctfbot-tools";
        paths = runtimeTools;
      };
    in
    {
      packages.${system}.default = pkgs.buildGoModule {
        pname = "ctfbot";
        version = "0.1.0";
        src = ./.;

        # Dependencies are vendored (see ./vendor), so no network/hash needed.
        vendorHash = null;

        # The sandbox integration tests exercise bwrap, which cannot run nested
        # inside the Nix build sandbox. Run them on a real host with `go test`.
        doCheck = false;

        nativeBuildInputs = [ pkgs.makeWrapper ];

        # add bwrap to PATH and SANDBOX_TOOLS to environment
        postInstall = ''
          wrapProgram $out/bin/ctfbot \
            --prefix PATH : ${pkgs.lib.makeBinPath ([ pkgs.bubblewrap ] ++ runtimeTools)} \
            --set SANDBOX_TOOLS ${sandboxTools}
        '';

        meta.description = "@ctfbot_ - Stacker News CTF bot";
      };

      apps.${system}.default = {
        type = "app";
        program = "${self.packages.${system}.default}/bin/ctfbot";
      };

      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [ go gopls bubblewrap ] ++ runtimeTools;
        SANDBOX_TOOLS = sandboxTools;
        shellHook = ''
          echo "ctfbot dev shell."
          echo "bwrap=$(command -v bwrap) git=$(command -v git) pdfinfo=$(command -v pdfinfo)"
        '';
      };
    };
}
