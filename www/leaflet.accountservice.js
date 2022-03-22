/* global L, config*/

L.Control.AccountService = L.Control.extend({
    options: {
        position: 'topleft'
    },

    // hardcode for now
    _icons: {
        'Bed': "Item_SimpleBed_Icon",
        "Broadsider": "KrakenShipNewIcon",
        "Carrack": "Galley_Icon",
        "Cog": "Galley_Icon",
        "Harrier": "Raft_Icon",
        "Kraken": "KrakenShipNewIcon",
        "MortarShip": "Raft_Icon",
        "Ramming_Galley": "Galley_Icon",
        "Galleon": "Item_GalleonHull_Icon",
        "Raft": "Raft_Icon",
        "Schooner": "Item_SchoonerHull_Icon",
        "Sloop": "Item_SloopHull_Icon",
        "Sloop_FromNPC": "Item_SloopHull_Icon",
        "Submarine": "Item_SloopHull_Icon",
        "TrampFreighter": "ICON_Tramp_Freighter",
        "TurtleShip": "Turtleship_Icon",
    },

    _ships: {},
    _eventSource: {},

    initialize: function(options) {
        L.Util.setOptions(this, options);
    },

    onAdd: function(map) {
        let container = L.DomUtil.create('div', 'leaflet-control-zoom leaflet-bar leaflet-control');
        this._map = map;
        fetch('/s/account', {
                dataType: 'json'
            }).then(r => {
                r.json().then(account => {
                    this._createButton('<i class="fa-solid fa-arrow-right-from-bracket"></i>', 'logout',
                        'leaflet-control-pin leaflet-bar-part leaflet-bar-part-top-and-bottom',
                        container, this._logout, this)
                    this._startEventListener(map);
                }).catch(error => {
                    this._createButton('<i class="fa-brands fa-steam" aria-hidden="true"></i>', 'Login with Steam',
                        'leaflet-control-pin leaflet-bar-part leaflet-bar-part-top-and-bottom',
                        container, this._login, this)
                })
            })
            .catch(error => {
                console.log("backend unavailable; not enabling login", error)
            });
        return container
    },

    onRemove: function(map) {

    },

    _startEventListener: function(map) {
        this._eventSource = new EventSource("/s/events");
        this._eventSource.onmessage = (event => {
            let d = JSON.parse(event.data);
            if (d.EntityType !== undefined) {
                this._processEntity(d)
            }
        });
    },

    _processEntity: function(d) {
        switch (d.EntityType) {
            case "Ship":
            case "ETribeEntityType::Ship":
                this._trackShip(d)
                break

            case "Bed":
            case "ETribeEntityType::Bed":
                console.dir(d) // slap bed on map
                break
        }
    },
    _trackShip: function(d) {
        // Get server grid reference.
        let duration = 5000,
            x = d.ServerId >> 16,
            y = d.ServerId & 0xFFFF,
            unrealX = config.GridSize * d.X + config.GridSize * x,
            unrealY = config.GridSize * d.Y + config.GridSize * y,
            gps = unrealToLeaflet(unrealX, unrealY);

        console.log(d)

        let ship = this._ships[d.EntityID];
        if (ship === undefined) {
            ship = L.Marker.movingMarker([gps], [duration]).addTo(this._map)
        }

        ship.addLatLng(gps, duration)
        ship.start()
        console.log(unrealX, unrealY, gps)
        this._ships[d.EntityID] = ship;
    },

    _login: function() {
        window.location = "/login";
    },

    _logout: function() {
        window.location = "/logout";
    },

    _createButton: function(html, title, className, container, fn, context) {
        let link = L.DomUtil.create('a', className, container)
        link.innerHTML = html
        link.href = '#'
        link.title = title

        L.DomEvent
            .on(link, 'click', L.DomEvent.stopPropagation)
            .on(link, 'click', L.DomEvent.preventDefault)
            .on(link, 'click', fn, context)
            .on(link, 'dbclick', L.DomEvent.stopPropagation)
        return link
    },

    draw: function(grids) {
        return this;
    }
});

L.control.accountControl = function(options) {
    return new L.Control.AccountService(options)
}

var accountControl;
L.Map.addInitHook(function() {
    this.accountControl = new L.Control.AccountService()
    accountControl = this.accountControl;
    this.addControl(this.accountControl)
})