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
        pythonEnv = pkgs.python313.withPackages (
          ps: with ps; [
            browser-cookie3
          ]
        );
      in
      {
        packages = {
          default = pkgs.buildGoModule {
            pname = "timer";
            version = "1";
            src = ./.;
            vendorHash = "sha256-La87Fu3YogYSe3FCR3W1Je4dwOpM7YIuMO+JNkASNYM=";
            # Tools needed at build-time (host)
            proxyVendor = true;
            nativeBuildInputs = [ pkgs.pkg-config ];

            # Libraries needed by the executable
            buildInputs = buildDeps;
          };
        };
        devShells = {
          default = pkgs.mkShell {
            buildInputs = [ pythonEnv ];

            # # Use 'inputsFrom' to pull dependencies from the package automatically
            # inputsFrom = [ self.packages.${system}.default ];

            # # Add extra development tools here
            # nativeBuildInputs = with pkgs; [
            #   go
            #   gopls
            # ];
          };
        };
      }
    );
}
