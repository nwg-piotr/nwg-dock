get:
	go get github.com/gotk3/gotk3@289cfb6dbf32de11dd2a392e86de4a144ac6be48
	go get github.com/gotk3/gotk3/gdk
	go get github.com/gotk3/gotk3/glib
	go get github.com/dlasky/gotk3-layershell/layershell
	go get github.com/joshuarubin/go-sway
	go get github.com/allan-simon/go-singleinstance

build:
	go build -o bin/nwg-dock *.go

install:
	cp bin/nwg-dock /usr/bin
	cp stuff/nwggrid.svg /usr/share/pixmaps

uninstall:
	rm /usr/bin/nwg-dock

run:
	go run *.go
