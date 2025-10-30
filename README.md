![A pixellated, fuzzy image that says "Shanty" on the left. On the right is a boat with two fishermen with fishing rods, one with a boot on the hook. Below is a swarm of fish.](/assets/images/logo.png)

A Navidrome music player for the Terminal, written in Go. Uses [MPV](https://github.com/mpv-player/mpv) for playing audio, [bubbletea](https://github.com/charmbracelet/bubbletea) for the TUI, and [chafa](https://github.com/hpjansson/chafa) for displaying images.
## Installation
### Requirements
- chafa (For images)
## Configuration
Currently configuration is done via a "config.toml" in the directory the program is run.
### Config Options

| Option | Type | Description |
| --- | --- | --- |
| serverUrl | String | The URL/IP that points to the navidrome server. |
| serverUser | String | Username of Navidrome user. |
| serverPassword | String | Password of Navidrome user.|

## Usage

| Focused | Keybind | Action |
| --- | --- | --- |
| N/A | ctrl+c | Exit shanty |
| N/A | shift+j | Move focus down |
| N/A | shift+k | Move focus up |
| Controls | Spacebar | Toggle play/pause |
| Controls | h | Go back 5 seconds |
| Controls | l | Go forward 5 seconds |
| Controls | j | Turn volume down by 5% |
| Controls | k | Turn volume up by 5% |
| Controls | n | Next track |
| Controls | p | Previous track |


