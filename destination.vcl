vcl 4.1;

backend destination {
	.host = "CHANGEME.api-server.example.com";
	.port = "443";
	.via = sslon;
}
