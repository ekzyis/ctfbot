{
  description = "ctfbot - a Stacker News shell CTF bot";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};

      # Packages that make up the runtime environment: they go on the bot's
      # PATH (so its startup command checks pass) and are bundled into
      # SANDBOX_TOOLS, the read-only command set mounted into the sandbox.
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
        vim # for `view`-style read-only poking; harmless in a ro fs
        pkgs."poppler-utils" # pdfinfo, pdftotext, ...
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

        # bwrap plus the curated command set go on PATH (the bot checks for a
        # few at startup); SANDBOX_TOOLS is the read-only tool dir it mounts.
        postInstall = ''
          wrapProgram $out/bin/ctfbot \
            --prefix PATH : ${pkgs.lib.makeBinPath ([ pkgs.bubblewrap ] ++ runtimeTools)} \
            --set SANDBOX_TOOLS ${sandboxTools}
        '';

        meta.description = "Stacker News bot for shell CTFs";
      };

      apps.${system}.default = {
        type = "app";
        program = "${self.packages.${system}.default}/bin/ctfbot";
      };

      devShells.${system}.default = pkgs.mkShell {
        # The same tools as production on PATH, plus SANDBOX_TOOLS set, so
        # `go run .` / `go test` see the identical runtime environment.
        packages = with pkgs; [ go gopls bubblewrap ] ++ runtimeTools;
        SANDBOX_TOOLS = sandboxTools;
        shellHook = ''
          echo "ctfbot dev shell. SANDBOX_TOOLS=$SANDBOX_TOOLS"
          echo "bwrap=$(command -v bwrap) git=$(command -v git) pdfinfo=$(command -v pdfinfo)"
        '';
      };
    };
}
