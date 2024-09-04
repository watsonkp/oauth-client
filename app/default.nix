let
        pkgs = import <nixpkgs> { };
        server = pkgs.callPackage ./server/default.nix { };
        container = server:
                pkgs.dockerTools.buildImage {
                        name = "sulliedsecurity/oauth-client";
                        tag = "testing";
                        copyToRoot = [
				server
				pkgs.dockerTools.caCertificates
			];
                        config = {
                                Cmd = [ "/bin/oauthclient" ];
                                Volumes = {
                                        "/static/" = {};
                                };
                        };
                };
in container server
