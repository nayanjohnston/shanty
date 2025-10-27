![A pixellated, fuzzy image that says "Shanty" on the left. On the right is a boat with two fishermen with fishing rods, one with a boot on the hook. Below is a swarm of fish.](/assets/images/logo.png)

A Navidrome music player for the Terminal, written in Go. Uses [MPV](https://github.com/mpv-player/mpv) for playing audio and [bubbletea](https://github.com/charmbracelet/bubbletea) for the TUI.
## Configuration
Currently configuration is done via a "config.toml" in the directory the program is run.
### Config Options

| Option | Type | Description |
| --- | --- | --- |
| serverUrl | String | The URL/IP that points to the navidrome server. |
| serverUser | String | Username of Navidrome user. |
| serverPassword | String | Password of Navidrome user.|


