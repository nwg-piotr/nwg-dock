package main

import (
    "fmt"
    "log"
    "path/filepath"

    "github.com/dlasky/gotk3-layershell/layershell"
    "github.com/gotk3/gotk3/gdk"
    "github.com/gotk3/gotk3/gtk"
)

var (
    appDirs         []string
    configDirectory string
)

func main() {
    configDirectory = configDir()
    // if doesn't exist
    createDir(configDirectory)
    cssFile := filepath.Join(configDirectory, "style.css")
    if pathExists(cssFile) {
        fmt.Printf("File %s found\n", cssFile)
    } else {
        fmt.Printf("File %s not found\n", cssFile)
    }

    appDirs = getAppDirs()

    tasks, err := listTasks()
    if err != nil {
        log.Fatal("Couldn't list tasks:", err)
    }

    for _, task := range tasks {
        fmt.Printf("%s on WS %v, PID %v, Name: '%s'\n", task.ID, task.WsNum, task.PID, task.Name)
    }

    gtk.Init(nil)

    cssProvider, err := gtk.CssProviderNew()
    if err != nil {
        fmt.Println(err)
    }
    err = cssProvider.LoadFromPath(cssFile)
    if err != nil {
        fmt.Println(err)
    }
    screen, _ := gdk.ScreenGetDefault()
    gtk.AddProviderForScreen(screen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_USER)

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
    layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_BOTTOM, 0)

    win.Connect("destroy", func() {
        gtk.MainQuit()
    })

    vbox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
    win.Add(vbox)

    hbox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
    vbox.PackStart(hbox, true, true, 0)

    for _, task := range tasks {
        button := createButton(task.ID, task.WsNum)
        hbox.PackStart(button, false, false, 0)
    }

    win.ShowAll()
    gtk.Main()
}
