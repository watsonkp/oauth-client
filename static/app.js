console.log('Hello, world!');

let STATE = {};

async function storeResourceOwnerAccessToken() {
	if (document.location.pathname == '/authorized') {
		let url = new URL(document.location.toString());
		url.pathname = '/token';
		const response = await fetch(url);
		if (response.ok) {
			const token_data = await response.json();
			STATE['access_token'] = token_data;
			history.replaceState(null, '', '/');
		}
	}
}

storeResourceOwnerAccessToken();
