# (Experimental) Atlas Map Viewer

## Setup
First:
1. Run the [Extract Plugin](https://github.com/antihax/ATLAS-Extract-Plugin) to obtain a JSON dump of your network.
2. Combine with the [ATLAS Map UI](https://github.com/antihax/atlasmap-js)
3. Edit the `json\config.js` and set the `AtlasMapServer` variable to the URL for this webservice.
3. Configure `ORIGIN_ALLOWED` environment variable to the URL of the server hosting the ATLAS Map UI content.

Then use one of the following host options:
### Independant Hosting
Use a webproxy like NGINX, APISIX, or HAProxy.
1. Route all `/` to a folder hosting the static ATLAS Map UI content.
2. Route `/s/*` and `/api` to the webservice.

### Internal Reverse Proxy
1. Host the static ATLAS Map UI content on an external webserver.
2. Set `STATICPROXY` environment variable to the URL of the external webserver.
3. Configure `ORIGIN_ALLOWED` to our URL.

### Internal Webserver
1. Set `STATICDIR` environment variable to the path containing the static ATLAS Map UI content.

### External Webserver
1. Host the static ATLAS Map UI content on an external server.
2. Edit the `json\config.js` and set the `AtlasMapServer` variable to the URL for this webservice.


# Web Service Configuration
The following environment variables can be set to reconfigure the service:

`PORT` webservice port. default 8880

`FETCHRATE` Atlas Redis polling frequency in seconds. default 15

`STATICDIR` location of static server files. default off

`STATICPROXY` proxy from external webserver for static server files. default off

`ORIGIN_ALLOWED` CORS allowed header. Should be set to the domain hosting this API.

`SESSION_PATH` location of session store files. default ./store

`SESSION_KEY` Session encryption key *MUST BE SET ON PRODUCTION* and should be a 32 byte value. default is random.

`ATLAS_REDIS_ADDRESS` Atlas Redis Address. default is localhost:6379.

`ATLAS_REDIS_PASSWORD` Atlas Redis Password. default is no password.

`ATLAS_REDIS_DB` Atlas Redis DB. default is 0.

`ADMIN_STEAMID_LIST` Space seperated list of Server Administrator SteamIDs. default is blank
