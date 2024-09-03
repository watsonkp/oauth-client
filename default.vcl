vcl 4.1;

# https://github.com/varnish/toolbox/blob/main/vcls/hit-miss/hit-miss.vcl
include "hit-miss.vcl";

import std;

backend default none;

backend sslon {
	.host = "127.0.0.1";
	.port = "8080";
}

include "destination.vcl";

sub vcl_recv {
	if (req.http.host == std.getenv("PROXY_FOR_HOST")) {
		set req.backend_hint = destination;
	} else {
		set req.backend_hint = default;
	}
}

sub vcl_miss {
	# Use the client access token if there is no user access token.
	if (!req.http.Authorization) {
		set req.http.Authorization = std.getenv("AUTHORIZATION");
	}
}

sub vcl_deliver {
	set resp.http.Access-Control-Allow-Origin = std.getenv("ALLOW_ORIGIN");
}

# https://github.com/varnish/toolbox/blob/main/vcls/verbose_builtin/verbose_builtin.vcl
include "verbose_builtin.vcl";
