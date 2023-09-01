PREFIX ?= /usr
DESTDIR ?=

get:
	go get github.com/gotk3/gotk3
	go get github.com/gotk3/gotk3/gdk
	go get github.com/gotk3/gotk3/glib
	go get github.com/dlasky/gotk3-layershell/layershell
	go get github.com/joshuarubin/go-sway
	go get github.com/allan-simon/go-singleinstance
	go get "github.com/sirupsen/logrus"

build:
	go build -o bin/nwg-dock .

install:
	-pkill -f nwg-dock
	sleep 1
	mkdir -p $(DESTDIR)$(PREFIX)/share/nwg-dock
	mkdir -p $(DESTDIR)$(PREFIX)/bin
	cp -r images $(DESTDIR)$(PREFIX)/share/nwg-dock
	cp config/* $(DESTDIR)$(PREFIX)/share/nwg-dock
	cp bin/nwg-dock $(DESTDIR)$(PREFIX)/bin

uninstall:
	rm -r $(DESTDIR)$(PREFIX)/share/nwg-dock
	rm $(DESTDIR)$(PREFIX)/bin/nwg-dock

run:
	go run .
