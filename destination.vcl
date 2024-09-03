vcl 4.1;

backend destination {
	.host = "change-me.api-server.example.com";
	.port = "443";
	.via = sslon;
}
