# nwg-dock

Fully configurable (w/ command line arguments and css) dock, written in Go, aimed exclusively at [sway](https://github.com/swaywm/sway) Wayland compositor. It features pinned buttons, task buttons, the workspace switcher and the launcher button. The latter by default starts `nwggrid` (application grid) from [nwg-launchers](https://github.com/nwg-piotr/nwg-launchers). In the picture(s) below the dock has been shown together with [nwg-panel](https://github.com/nwg-piotr/nwg-panel).

![06.png](https://scrot.cloud/images/2021/04/02/06.png)

[more pictures](https://scrot.cloud/album/nwg-dock.BuZM)

## Installation

### Requirements

- `go` up to 1.16.2: just to build. **Please note**, that the gtk-layer-shell library does not yet seem to behave well with **go 1.16.3**.
- `gtk3`
- `gtk-layer-shell`
- `nwg-launchers`: optionally. You may use another launcher, see help.


### Steps

1. Clone the repository, cd into it.
2. Install necessary golang libraries with `make get`. First time it may take awhile.
3. `sudo make install`

Or you may try just `sudo make install`, to install the binary you downloaded in the `/bin` directory.

## Running

Either start the dock permanently in the sway config file, or assign the command to some key binding.
Running the command again kills existing program instance, so you may use the same key to open and close the dock.

```txt
Usage of nwg-dock:
  -a string
    	Alignment in full width/height: "start", "center" or "end" (default "center")
  -c string
    	Command assigned to the launcher button (default "nwggrid -p")
  -d	auto-hiDe: close window when left or a button clicked
  -f	take Full screen width/height
  -i int
    	Icon size (default 48)
  -l string
    	Layer "top" or "bottom" (default "top")
  -mb int
    	Margin Bottom
  -ml int
    	Margin Left
  -mr int
    	Margin Right
  -mt int
    	Margin Top
  -o string
    	name of Output to display the dock on
  -p string
    	Position: "bottom", "top" or "left" (default "bottom")
  -s string
    	Styling: css file name (default "style.css")
  -v	display Version information
  -w int
    	number of Workspaces you use (default 8)
  -x	set eXclusive zone: move other windows aside

```

## Styling

Edit `~/.config/nwg-dock/style.css` to your taste.

## Credits

This program uses some great libraries:

- [gotk3](https://github.com/gotk3/gotk3) Copyright (c) 2013-2014 Conformal Systems LLC,
Copyright (c) 2015-2018 gotk3 contributors
- [gotk3-layershell](https://github.com/dlasky/gotk3-layershell) by [@dlasky](https://github.com/dlasky/gotk3-layershell/commits?author=dlasky) - many thanks for writing this software, and for patience with my requests!
- [go-sway](https://github.com/joshuarubin/go-sway) Copyright (c) 2019 Joshua Rubin
- [go-singleinstance](github.com/allan-simon/go-singleinstance) Copyright (c) 2015 Allan Simon
