vcl 4.1;

# https://github.com/varnish/toolbox/blob/main/vcls/hit-miss/hit-miss.vcl
include "hit-miss.vcl";

import std;

backend default {
	.host = "127.0.0.1";
	.port = "8081";
}

backend sslon {
	.host = "127.0.0.1";
	.port = "8080";
}

include "destination.vcl";

sub vcl_recv {
	# Never cache state tokens (CSRF protection)
	if (req.url == "/state") {
		return(pass);
	} elif (req.url ~ "^/api/") {
		set req.url = regsub(req.url, "^/api/", "/");
		set req.backend_hint = destination;
	} else {
		set req.backend_hint = default;
	}
}

sub vcl_miss {
	# Use the client access token if there is no user access token.
	if ((req.backend_hint == destination) && (!req.http.Authorization)) {
		set req.http.Authorization = std.getenv("AUTHORIZATION");
	}
}

sub vcl_backend_response {
	if ((bereq.url == "/") || (bereq.url == "/index.html") || (bereq.url ~ "^/authorized")) {
		set beresp.do_esi = true;
	}
}

# https://github.com/varnish/toolbox/blob/main/vcls/verbose_builtin/verbose_builtin.vcl
include "verbose_builtin.vcl";
