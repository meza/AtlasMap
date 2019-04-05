# (Experimental) Atlas Map Viewer
![Alt text](Example1.jpg?raw=true "Exmaple1")
Used as an example to read ship and bed positions from Redis overlayed on top of a world and territory map.  It consists of two pieces, a go web service that retrieves and serves the ship and bed positions and a simple [React](https://reactjs.org/) / [Leaflet.js](https://leafletjs.com/) app.

The slippy map tiles for the world are generated by [ServerGridEditor](https://github.com/GrapeshotGames/ServerGridEditor) and should be placed in the "www/tiles" directory.

## Setup
You need to generate a slippy style map (See ServerGridEditor or other tools), clear out and replace all files in `www/tiles/*` with your map files. Make sure to match the server grid size in config.js and setup the following config options in sections below. Note: Untested with non square grids currently but we definitely want this to work if it doesn't already

### Web Service
The following environment variables can be set to reconfigure the service:

	`PORT` webservice port. default 8880
	`DISABLECOMMANDS` disables the remote commands to the cluster. default true
	`FETCHRATE` Atlas Redis polling frequency in seconds. default 15
    `TERRITORY_URL` Optional link to territory server. default http://localhost:8881/territoryTiles/
    `STATICDIR` location of static server files. default ./www
    `SESSION_PATH` location of session store files. default ./store
    `SESSION_KEY` Session encryption key *MUST BE SET ON PRODUCTION*. default is random.

    `ATLAS_REDIS_ADDRESS` Atlas Redis Address. default is localhost:6379.
    `ATLAS_REDIS_PASSWORD` Atlas Redis Password. default is foobared.
    `ATLAS_REDIS_DB` Atlas Redis DB. default is 0.

### Web App
The client file "www/config.js" holds some cluster specific information like the grid size.
```
const config = {
    //Number of columns in the grid
    ServersX: 15,
	
    // Number of rows in the grid
    ServersY: 15,
	
    //Command completion suggestion
    Suggestions: [
        "spawnshipfast name",
        "spawnbed name",
        "destroyall",
    ]
}
```

