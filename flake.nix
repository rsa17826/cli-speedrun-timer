{
  inputs = {
    nixpkgs = {
      url = "github:NixOS/nixpkgs/nixos-unstable";
    };
    flake-utils = {
      url = "github:numtide/flake-utils";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };

        # Define the dependencies once so you don't repeat yourself
        buildDeps = with pkgs; [
          libGL
          libX11
          libXrandr
          libXinerama
          libXcursor
          libXi
          libXxf86vm
          mesa
        ];
      in
      {
        packages = {
          default = pkgs.buildGoModule {
            pname = "timer";
            version = "1";
            src = ./.;
            vendorHash = "sha256-eWOvslPtMStQAxk9YU+uXUObxiLY7pWzvU0ffUcMx+E=";
            # Tools needed at build-time (host)
            proxyVendor = true;
            nativeBuildInputs = [ pkgs.pkg-config ];

            # Libraries needed by the executable
            buildInputs = buildDeps;
          };
        };
        devShells = {
          default = pkgs.mkShell {
            # Use 'inputsFrom' to pull dependencies from the package automatically
            inputsFrom = [ self.packages.${system}.default ];

            # Add extra development tools here
            nativeBuildInputs = with pkgs; [
              go
              gopls
            ];
          };
        };
      }
    );
}
