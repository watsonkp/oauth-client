{ pkgs ? import <nixpkgs> {} }:

pkgs.buildGoModule rec {
        pname = "oauthclient";
        version = "0.1.0";

        src = ./.;

        vendorHash = null;

        meta = with pkgs.lib; {
                description = "OAuth 2.0 client application";
                homepage = "https://git.sulliedsecurity.com/kevin/oauth-client";
                license = licenses.mit;
                maintainers = with maintainers; [ watsonkp ];
        };
}
