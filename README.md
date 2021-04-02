# nwg-dock

Fully configurable (w/ command line arguments and css) dock, written in Go, aimed exclusively at sway Wayland compositor.

## Installation

### Requirements

- Go 1.16
- `nwg-launchers` package installed: optionally. You may use another launcher, see help.

### Steps

1. Clone the repository, cd into it.
2. Install necessary golang libraries with `make get`. First time it may take awhile.
3. `sudo make install`

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
