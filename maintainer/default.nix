let
        pkgs = import <nixpkgs> { };
        agent = pkgs.callPackage ./agent/default.nix { };
        container = agent:
                pkgs.dockerTools.buildImage {
                        name = "sulliedsecurity/oauth-maintainer";
                        tag = "testing";
                        copyToRoot = [
				agent
				pkgs.dockerTools.caCertificates
			];
                        config = {
                                Cmd = [ "/bin/agent" ];
                        };
                };
in container agent
