package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gtk"

	"github.com/joshuarubin/go-sway"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	client, err := sway.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	tree, _ := client.GetTree(ctx)

	nodes := tree.Nodes

	var outputs []*sway.Node
	for _, n := range nodes {
		if n.Type == "output" && !strings.HasPrefix(n.Name, "__") {
			outputs = append(outputs, n)
		}
	}

	var workspaces []*sway.Node
	for _, o := range outputs {
		oNodes := o.Nodes
		for _, n := range oNodes {
			if n.Type == "workspace" {
				workspaces = append(workspaces, n)
			}
		}
	}
	for _, w := range workspaces {
		wNodes := w.Nodes
		fmt.Printf("Workspace %s:\n", w.Name)
		for _, con := range wNodes {
			if con.AppID != nil {
				fmt.Println(*con.AppID, con.Name, *con.PID)
			} else {
				wp := *con.WindowProperties
				fmt.Println(wp.Class, con.Name)
			}
		}
	}

	// Initialize GTK without parsing any command line arguments.
	gtk.Init(nil)

	// Create a new toplevel window, set its title, and connect it to the
	// "destroy" signal to exit the GTK main loop when it is destroyed.
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	layershell.InitForWindow(win)

	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_LEFT, false)
	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_BOTTOM, true)
	layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_RIGHT, false)

	layershell.SetLayer(win, layershell.LAYER_SHELL_LAYER_TOP)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_TOP, 0)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_LEFT, 0)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_RIGHT, 0)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_BOTTOM, 6)

	win.SetTitle("Simple Example")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	// Create a new label widget to show in the window.
	l, err := gtk.LabelNew("Hello, gotk3!")
	if err != nil {
		log.Fatal("Unable to create label:", err)
	}

	// Add the label to the window.
	win.Add(l)

	// Set the default window size.
	win.SetDefaultSize(800, 30)

	// Recursively show all widgets contained in this window.
	win.ShowAll()

	// Begin executing the GTK main loop.  This blocks until
	// gtk.MainQuit() is run.
	gtk.Main()
}
