package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/allan-simon/go-singleinstance"
	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const version = "0.1.4"

var (
	appDirs                            []string
	dataHome                           string
	configDirectory                    string
	pinnedFile                         string
	pinned                             []string
	oldTasks                           []task
	mainBox                            *gtk.Box
	src                                glib.SourceHandle
	refresh                            bool // we will use this to trigger rebuilding mainBox
	outerOrientation, innerOrientation gtk.Orientation
	widgetAnchor, menuAnchor           gdk.Gravity
	imgSizeScaled                      int
	currentWsNum, targetWsNum          int64
	dockWindow                         *gtk.Window
)

// Flags
var cssFileName = flag.String("s", "style.css", "Styling: css file name")
var targetOutput = flag.String("o", "", "name of Output to display the dock on")
var displayVersion = flag.Bool("v", false, "display Version information")
var autohide = flag.Bool("d", false, "auto-hiDe: show dock when hotspot hovered, close when left or a button clicked")
var full = flag.Bool("f", false, "take Full screen width/height")
var numWS = flag.Int64("w", 8, "number of Workspaces you use")
var position = flag.String("p", "bottom", "Position: \"bottom\", \"top\" or \"left\"")
var exclusive = flag.Bool("x", false, "set eXclusive zone: move other windows aside; overrides the \"-l\" argument")
var imgSize = flag.Int("i", 48, "Icon size")
var layer = flag.String("l", "overlay", "Layer \"overlay\", \"top\" or \"bottom\"")
var launcherCmd = flag.String("c", "", "Command assigned to the launcher button")
var alignment = flag.String("a", "center", "Alignment in full width/height: \"start\", \"center\" or \"end\"")
var marginTop = flag.Int("mt", 0, "Margin Top")
var marginLeft = flag.Int("ml", 0, "Margin Left")
var marginRight = flag.Int("mr", 0, "Margin Right")
var marginBottom = flag.Int("mb", 0, "Margin Bottom")
var noWs = flag.Bool("nows", false, "don't show the workspace switcher")
var noLauncher = flag.Bool("nolauncher", false, "don't show the launcher button")

func buildMainBox(tasks []task, vbox *gtk.Box) {
	mainBox.Destroy()
	mainBox, _ = gtk.BoxNew(innerOrientation, 0)

	if *alignment == "start" {
		vbox.PackStart(mainBox, false, true, 0)
	} else if *alignment == "end" {
		vbox.PackEnd(mainBox, false, true, 0)
	} else {
		vbox.PackStart(mainBox, true, false, 0)
	}

	var err error
	pinned, err = loadTextFile(pinnedFile)
	if err != nil {
		pinned = nil
	}

	var allItems []string
	for _, cntPin := range pinned {
		if !isIn(allItems, cntPin) {
			allItems = append(allItems, cntPin)
		}
	}
	for _, cntTask := range tasks {
		if !isIn(allItems, cntTask.ID) && !strings.Contains(*launcherCmd, cntTask.ID) {
			allItems = append(allItems, cntTask.ID)
		}
	}

	// scale icons down when their number increases
	if *imgSize*6/(len(allItems)) < *imgSize {
		overflow := (len(allItems) - 6) / 3
		imgSizeScaled = *imgSize * 6 / (6 + overflow)
	} else {
		imgSizeScaled = *imgSize
	}

	var alreadyAdded []string
	for _, pin := range pinned {
		if !inTasks(tasks, pin) {
			button := pinnedButton(pin)
			mainBox.PackStart(button, false, false, 0)
		} else {
			instances := taskInstances(pin, tasks)
			task := instances[0]
			if len(instances) == 1 {
				button := taskButton(task, instances)
				mainBox.PackStart(button, false, false, 0)
			} else if !isIn(alreadyAdded, task.ID) {
				button := taskButton(task, instances)
				mainBox.PackStart(button, false, false, 0)
				alreadyAdded = append(alreadyAdded, task.ID)
				taskMenu(task.ID, instances)
			} else {
				continue
			}
		}
	}

	alreadyAdded = nil
	for _, task := range tasks {
		// nwggrid is the default launcher, we don't want to see it as a task
		if !inPinned(task.ID) && task.ID != "nwggrid" {
			instances := taskInstances(task.ID, tasks)
			if len(instances) == 1 {
				button := taskButton(task, instances)
				mainBox.PackStart(button, false, false, 0)
			} else if !isIn(alreadyAdded, task.ID) {
				button := taskButton(task, instances)
				mainBox.PackStart(button, false, false, 0)
				alreadyAdded = append(alreadyAdded, task.ID)
				taskMenu(task.ID, instances)
			} else {
				continue
			}
		}
	}

	if !*noWs {
		wsButton, _ := gtk.ButtonNew()
		wsImage, err := createImage(filepath.Join(dataHome, fmt.Sprintf("nwg-dock/images/%v.svg", currentWsNum)),
			imgSizeScaled)
		if err == nil {
			wsButton.SetImage(wsImage)
			wsButton.SetAlwaysShowImage(true)
			wsButton.AddEvents(int(gdk.SCROLL_MASK))

			wsButton.Connect("clicked", func() {
				focusWorkspace(targetWsNum)
			})

			wsButton.Connect("enter-notify-event", cancelClose)

			wsButton.Connect("scroll-event", func(btn *gtk.Button, e *gdk.Event) bool {
				event := gdk.EventScrollNewFromEvent(e)
				if event.Direction() == gdk.SCROLL_UP {
					if targetWsNum < *numWS && targetWsNum < 20 {
						targetWsNum++
					} else {
						targetWsNum = 1
					}
					pixbuf, _ := gdk.PixbufNewFromFileAtSize(filepath.Join(dataHome, fmt.Sprintf("nwg-dock/images/%v.svg",
						targetWsNum)), imgSizeScaled, imgSizeScaled)
					wsImage.SetFromPixbuf(pixbuf)

					return true
				} else if event.Direction() == gdk.SCROLL_DOWN {
					if targetWsNum > 1 {
						targetWsNum--
					} else {
						targetWsNum = *numWS
					}
					pixbuf, _ := gdk.PixbufNewFromFileAtSize(filepath.Join(dataHome, fmt.Sprintf("nwg-dock/images/%v.svg",
						targetWsNum)), imgSizeScaled, imgSizeScaled)
					wsImage.SetFromPixbuf(pixbuf)

					return true
				}
				return false
			})
		}
		mainBox.PackStart(wsButton, false, false, 0)
	}

	if !*noLauncher && *launcherCmd != "" {
		button, _ := gtk.ButtonNew()
		image, err := createImage(filepath.Join(dataHome, "nwg-dock/images/grid.svg"), imgSizeScaled)
		if err == nil {
			button.SetImage(image)
			button.SetAlwaysShowImage(true)

			button.Connect("clicked", func() {
				elements := strings.Split(*launcherCmd, " ")
				cmd := exec.Command(elements[0], elements[1:]...)
				go cmd.Run()
				if *autohide {
					dockWindow.Hide()
				}
			})
			button.Connect("enter-notify-event", cancelClose)
		}
		mainBox.PackStart(button, false, false, 0)
	}

	mainBox.ShowAll()
}

func setupHotSpot(monitor gdk.Monitor, dockWindow *gtk.Window) gtk.Window {
	w, h := dockWindow.GetSize()

	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)

	layershell.InitForWindow(win)
	layershell.SetMonitor(win, &monitor)

	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	win.Add(box)

	win.Connect("enter-notify-event", func() {
		layershell.SetMonitor(dockWindow, &monitor)
		dockWindow.Hide()
		dockWindow.Show()
	})

	if *position == "bottom" || *position == "top" {
		win.SetSizeRequest(w, 10)
		if *position == "bottom" {
			layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_BOTTOM, true)
		} else {
			layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_TOP, true)
		}

		layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_LEFT, *full)
		layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_RIGHT, *full)
	}

	if *position == "left" {
		win.SetSizeRequest(10, h)
		layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_LEFT, true)

		layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_TOP, *full)
		layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_BOTTOM, *full)
	}

	layershell.SetLayer(win, layershell.LAYER_SHELL_LAYER_TOP)

	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_TOP, *marginTop)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_LEFT, *marginLeft)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_RIGHT, *marginRight)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_BOTTOM, *marginBottom)

	layershell.SetExclusiveZone(win, -1)

	return *win
}

func main() {
	flag.Parse()

	if *displayVersion {
		fmt.Printf("nwg-dock version %s\n", version)
		os.Exit(0)
	}

	// Gentle SIGTERM handler thanks to reiki4040 https://gist.github.com/reiki4040/be3705f307d3cd136e85
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	go func() {
		for {
			s := <-signalChan
			if s == syscall.SIGTERM {
				println("SIGTERM received, bye bye!")
				gtk.MainQuit()
			}
		}
	}()

	// Unless we are in autohide mode, we probably want the same key/mouse binding to turn the dock off.
	// Kill the running instance and exit.
	lockFilePath := fmt.Sprintf("%s/nwg-dock.lock", tempDir())
	lockFile, err := singleinstance.CreateLockFile(lockFilePath)
	if err != nil {
		pid, err := readTextFile(lockFilePath)
		if err == nil {
			i, err := strconv.Atoi(pid)
			if err == nil {
				if !*autohide {
					println("Running instance found, sending SIGTERM and exiting...")
					syscall.Kill(i, syscall.SIGTERM)
				} else {
					println("Already running")
				}
			}
		}
		os.Exit(0)
	}
	defer lockFile.Close()

	if !*noLauncher && *launcherCmd == "" {
		if isCommand("nwg-drawer") {
			*launcherCmd = "nwg-drawer"
		} else if isCommand("nwggrid") {
			*launcherCmd = "nwggrid -p"
		}

		if *launcherCmd != "" {
			println(fmt.Sprintf("Using auto-detected launcher command: '%s'", *launcherCmd))
		} else {
			println("Neither 'nwg-drawer' nor 'nwggrid' command found, and no other launcher specified; hiding the launcher button.")
		}
	}

	dataHome = getDataHome()
	configDirectory = configDir()
	// if doesn't exist:
	createDir(configDirectory)

	if !pathExists(fmt.Sprintf("%s/style.css", configDirectory)) {
		copyFile(filepath.Join(dataHome, "nwg-dock/style.css"), fmt.Sprintf("%s/style.css", configDirectory))
	}

	cacheDirectory := cacheDir()
	if cacheDirectory == "" {
		log.Panic("Couldn't determine cache directory location")
	}
	pinnedFile = filepath.Join(cacheDirectory, "nwg-dock-pinned")
	cssFile := filepath.Join(configDirectory, *cssFileName)

	appDirs = getAppDirs()

	gtk.Init(nil)

	cssProvider, _ := gtk.CssProviderNew()

	err = cssProvider.LoadFromPath(cssFile)
	if err != nil {
		fmt.Printf("%s file not found, using GTK styling\n", cssFile)
	} else {
		fmt.Printf("Using style: %s\n", cssFile)
		screen, _ := gdk.ScreenGetDefault()
		gtk.AddProviderForScreen(screen, cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
	}

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	dockWindow = win

	layershell.InitForWindow(win)

	var output2mon map[string]*gdk.Monitor
	if *targetOutput != "" {
		// We want to assign layershell to a monitor, but we only know the output name!
		output2mon, err = mapOutputs()
		if err == nil {
			layershell.SetMonitor(win, output2mon[*targetOutput])
		} else {
			println(err)
		}
	}

	if *exclusive {
		layershell.AutoExclusiveZoneEnable(win)
		*layer = "top"
	}

	if *position == "bottom" || *position == "top" {
		if *position == "bottom" {
			layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_BOTTOM, true)

			widgetAnchor = gdk.GDK_GRAVITY_NORTH
			menuAnchor = gdk.GDK_GRAVITY_SOUTH
		} else {
			layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_TOP, true)

			widgetAnchor = gdk.GDK_GRAVITY_SOUTH
			menuAnchor = gdk.GDK_GRAVITY_NORTH
		}

		outerOrientation = gtk.ORIENTATION_VERTICAL
		innerOrientation = gtk.ORIENTATION_HORIZONTAL

		layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_LEFT, *full)
		layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_RIGHT, *full)
	}

	if *position == "left" {
		layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_LEFT, true)

		layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_TOP, *full)
		layershell.SetAnchor(win, layershell.LAYER_SHELL_EDGE_BOTTOM, *full)

		outerOrientation = gtk.ORIENTATION_HORIZONTAL
		innerOrientation = gtk.ORIENTATION_VERTICAL

		widgetAnchor = gdk.GDK_GRAVITY_EAST
		menuAnchor = gdk.GDK_GRAVITY_WEST
	}

	if *autohide {
		// we need to cover the hotspot window, to avoid unwanted .Hide() and .Show() the dockWindow
		layershell.SetLayer(win, layershell.LAYER_SHELL_LAYER_OVERLAY)
		layershell.SetExclusiveZone(win, -1)
	} else {
		// otherwise let's leave users freedom of choice
		if *layer == "top" {
			layershell.SetLayer(win, layershell.LAYER_SHELL_LAYER_TOP)
		} else if *layer == "bottom" {
			layershell.SetLayer(win, layershell.LAYER_SHELL_LAYER_BOTTOM)
		} else {
			layershell.SetLayer(win, layershell.LAYER_SHELL_LAYER_OVERLAY)
			layershell.SetExclusiveZone(win, -1)
		}
	}

	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_TOP, *marginTop)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_LEFT, *marginLeft)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_RIGHT, *marginRight)
	layershell.SetMargin(win, layershell.LAYER_SHELL_EDGE_BOTTOM, *marginBottom)

	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	// Close the window on leave, but not immediately, to avoid accidental closes
	win.Connect("leave-notify-event", func() {
		if *autohide {
			src = glib.TimeoutAdd(uint(1000), func() bool {
				win.Hide()
				src = 0
				return false
			})
		}
	})

	win.Connect("enter-notify-event", func() {
		cancelClose()
	})

	outerBox, _ := gtk.BoxNew(outerOrientation, 0)
	outerBox.SetProperty("name", "box")
	win.Add(outerBox)

	alignmentBox, _ := gtk.BoxNew(innerOrientation, 0)
	outerBox.PackStart(alignmentBox, true, true, 0)

	mainBox, _ = gtk.BoxNew(innerOrientation, 0)
	// We'll pack mainBox later, in buildMainBox

	tasks, err := listTasks()
	if err != nil {
		log.Fatal("Couldn't list tasks:", err)
	}
	oldTasks = tasks
	var oldWsNum int64

	buildMainBox(tasks, alignmentBox)

	glib.TimeoutAdd(uint(150), func() bool {
		currentTasks, _ := listTasks()
		if len(currentTasks) != len(oldTasks) || currentWsNum != oldWsNum || refresh {
			println("refreshing...")
			buildMainBox(currentTasks, alignmentBox)
			oldTasks = currentTasks
			oldWsNum = currentWsNum
			targetWsNum = currentWsNum
			refresh = false
		}
		return true
	})

	win.ShowAll()

	if *autohide {
		win.Hide()

		mRefProvider, _ := gtk.CssProviderNew()
		if err := mRefProvider.LoadFromPath(filepath.Join(dataHome, "nwg-dock/hotspot.css")); err != nil {
			println(err)
		}

		if *targetOutput == "" {
			// hot spots on all displays
			monitors, _ := listMonitors()
			for _, monitor := range monitors {
				win := setupHotSpot(monitor, win)

				context, _ := win.GetStyleContext()
				context.AddProvider(mRefProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

				win.ShowAll()
			}
		} else {
			// hot spot on the selected display only
			monitor := output2mon[*targetOutput]
			win := setupHotSpot(*monitor, win)

			context, _ := win.GetStyleContext()
			context.AddProvider(mRefProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

			win.ShowAll()
		}
	}

	gtk.Main()
}
