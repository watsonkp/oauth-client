# OAuth Client

## Purpose
A Kubernetes web application architecture for interacting with registered external OAuth 2.0 APIs.

An application could be served as static content with client-side JavaScript handling the API authorization, but this would expose the client application API secrets to users. This is commonly mitigated by registering an allowed `referer` header with the API, monitoring usage, and revoking tokens.

This project is an alternative that limits secret access to the application server.

## Architecture
A Varnish cache which proxies traffic to an HAProxy backend that initiates TLS connections to the external API.

The Varnish cache injects the OAuth client access token when a user access token has not been provided.

Inbound TLS is handled by the Kubernetes ingress.

## Build
	cd app
	docker load <$(nix-build)
	docker tag sulliedsecurity/oauthclient:testing CHANGEME.registry.example.com/sulliedsecurity/oauthclient:testing
	docker push CHANGEME.registry.example.com/sulliedsecurity/oauthclient:testing

## Configure
1. Change the `.host` property in `destination.vcl` to the domain name of the API server.
2. Change the `server` property in `haproxy.cfg` to the API server.
4. In the `kustomization.yaml` file change the `access-token` value of the `client-authorization` configuration to an OAuth client access token.
5. Change the replacement registry values in `set_registry.yaml` for the Varnish and HAProxy images.

Check with `grep -r CHANGEME .`.

## Inspiration
James Kettle's research into web cache vulnerabilities motivated me to gain practical experience with web cache's and Varnish.
