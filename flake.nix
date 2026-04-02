{
  description = "Picowatcher - RSS feed monitor with LLM summarization";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go toolchain (use latest stable Go available in nixpkgs)
            go

            # Go development tools
            gopls          # Language server
            gotools        # Additional tools (goimports, etc)
            go-tools       # Static analysis tools
            delve          # Debugger

            # Build and development utilities
            git
            gnumake
          ];

          shellHook = ''
            echo "🚀 Picowatcher Go development environment"
            echo "Go version: $(go version)"
            echo ""
            echo "Available commands:"
            echo "  go build          - Build the application"
            echo "  go run main.go    - Run the application"
            echo "  go test ./...     - Run tests"
            echo "  go mod tidy       - Tidy dependencies"
            echo ""
          '';
        };

        # Optional: Define the package itself
        packages.default = pkgs.buildGoModule {
          pname = "picowatcher";
          version = "0.1.0";
          src = ./.;

          vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";

          meta = with pkgs.lib; {
            description = "RSS feed monitor with LLM summarization and webhook notifications";
            homepage = "https://github.com/j4ger/picowatcher";
            license = licenses.mit;
            maintainers = [ ];
          };
        };
      }
    );
}
