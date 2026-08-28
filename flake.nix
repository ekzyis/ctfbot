{
  description = "ctfbot - a Stacker News shell CTF bot";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};

      # The set of commands available to players on PATH inside the sandbox.
      # Add or remove packages here to change what the CTF box "has installed".
      sandboxTools = pkgs.buildEnv {
        name = "ctfbot-tools";
        paths = with pkgs; [
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
          vim # for `view`-style read-only poking; harmless in a ro fs
          pkgs."poppler-utils" # pdfinfo
        ];
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

        # bwrap must be on PATH at runtime; SANDBOX_TOOLS points the bot at the
        # curated command set above.
        postInstall = ''
          wrapProgram $out/bin/ctfbot \
            --prefix PATH : ${pkgs.lib.makeBinPath [ pkgs.bubblewrap ]} \
            --set-default SANDBOX_TOOLS ${sandboxTools}
        '';

        meta.description = "Stacker News bot for shell CTFs";
      };

      apps.${system}.default = {
        type = "app";
        program = "${self.packages.${system}.default}/bin/ctfbot";
      };

      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [ go gopls bubblewrap ];
        # So `go run .` / `go test` use the same curated tool set as production.
        SANDBOX_TOOLS = sandboxTools;
        shellHook = ''
          echo "ctfbot dev shell. SANDBOX_TOOLS=$SANDBOX_TOOLS"
          echo "bwrap: $(command -v bwrap)"
        '';
      };
    };
}
