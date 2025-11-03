<p align="center"><img src="/assets/images/logo.png"></p>

<p align="center">A terminal music player for <a href="https://github.com/navidrome/navidrome">Navidrome</a>, written in Go + <a href="https://github.com/charmbracelet/bubbletea">Bubble Tea</a>! Includes album art in the terminal via <a href="https://github.com/hpjansson/chafa">chafa</a>.</p>

<p align="center"><img src="/assets/images/screenshot-01.jpeg"></p>

# Notes

- This is the first (actual) program I've written in Go, so the code is a nightmare and stability isn't guaranteed.
- It currently only targets Linux (let's be honest, who else would want a music player in their terminal).

# Installation

## Requirements

To run shanty, you'll also need to install the following:
- [chafa](https://github.com/hpjansson/chafa) (For image display)
- [mpv](https://github.com/mpv-player/mpv) (For audio playback)

# Configuration

Configuration is required to run the program, and needs to be located at `~/.config/shanty/config.toml`.

## Config Options

| Option | Type | Description |
| --- | --- | --- |
| `serverUrl` | String | The URL/IP that points to the navidrome server. |
| `serverUser` | String | Username of Navidrome user. |
| `serverPassword` | String | Password of Navidrome user.|

# Usage

## Global

| Keybind | Action |
| --- | --- |
| <kbd>Ctrl+c</kbd> | Exit shanty |
| <kbd>Shift+j</kbd> | Move focus down |
| <kbd>Shift+k</kbd> | Move focus up |

## Controller

| Keybind | Action |
| --- | --- |
| <kbd>Spacebar</kbd> | Toggle play/pause |
| <kbd>h</kbd> | Go back 5 seconds |
| <kbd>l</kbd> | Go forward 5 seconds |
| <kbd>j</kbd> | Turn volume down by 5% |
| <kbd>k</kbd> | Turn volume up by 5% |
| <kbd>n</kbd> | Next track |
| <kbd>p</kbd> | Previous track |

## Library

| Keybind | Action |
| --- | --- |
| <kbd>h</kbd> <kbd>j</kbd> <kbd>k</kbd> <kbd>l</kbd> | Move selection |
| <kbd>n</kbd> | Next page |
| <kbd>p</kbd> | Previous page |
| <kbd>Enter</kbd> | Select album |

## Queue

| Keybind | Action |
| --- | --- |
| <kbd>j</kbd> <kbd>k</kbd> | Move selection |
| <kbd>Enter</kbd> | Play selected song |
| <kbd>d</kbd> | Remove selected song |

## Album

| Keybind | Action |
| --- | --- |
| <kbd>h</kbd> <kbd>j</kbd> <kbd>k</kbd> <kbd>l</kbd> | Move selection |
| <kbd>Enter</kbd> | Select current option (Selecting songs will add them to queue) |

# Packages Used

- [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- [Bubbles](https://github.com/charmbracelet/bubbles)
- [go-mpv](https://github.com/gen2brain/go-mpv)
- [go-toml v2](https://github.com/pelletier/go-toml)
