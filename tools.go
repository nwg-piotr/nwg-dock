package main

import (
    "context"
    "errors"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "time"

    "github.com/gotk3/gotk3/gdk"
    "github.com/gotk3/gotk3/gtk"
    "github.com/joshuarubin/go-sway"
)

var descendants []sway.Node

type task struct {
    conID int64
    ID    string // will be created out of app_id or window class
    Name  string
    PID   uint32
    WsNum int64
}

func taskInstances(ID string, tasks []task) []task {
    var found []task
    for _, t := range tasks {
        if t.ID == ID {
            found = append(found, t)
        }
    }
    return found
}

// list sway tree, return tasks sorted by workspace numbers
func listTasks() ([]task, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    client, err := sway.New(ctx)
    if err != nil {
        return nil, err
    }

    tree, err := client.GetTree(ctx)
    if err != nil {
        return nil, err
    }

    workspaces, _ := client.GetWorkspaces(ctx)
    if err != nil {
        return nil, err
    }

    // all nodes in the tree
    nodes := tree.Nodes

    // find outputs in all nodes
    var outputs []*sway.Node
    for _, n := range nodes {
        if n.Type == "output" && !strings.HasPrefix(n.Name, "__") {
            outputs = append(outputs, n)
        }
    }

    // find workspaces in outputs
    var workspaceNodes []*sway.Node
    for _, o := range outputs {
        nodes = o.Nodes
        for _, n := range nodes {
            if n.Type == "workspace" {
                workspaceNodes = append(workspaceNodes, n)
            }
        }
    }

    var tasks []task
    // find cons in workspaces recursively
    for _, w := range workspaceNodes {
        wsNum := workspaceNum(workspaces, w.Name)
        descendants = nil
        for _, con := range w.Nodes {
            findDescendants(*con)
        }

        // create tasks from cons which represent tasks
        for _, con := range descendants {
            tasks = append(tasks, createTask(con, wsNum))
        }

        fNodes := w.FloatingNodes
        for _, con := range fNodes {
            tasks = append(tasks, createTask(*con, wsNum))
        }

    }
    sort.Slice(tasks, func(i int, j int) bool {
        return tasks[i].WsNum < tasks[j].WsNum
    })
    return tasks, nil
}

func findDescendants(con sway.Node) {
    if len(con.Nodes) > 0 {
        for _, node := range con.Nodes {
            findDescendants(*node)
        }
    } else {
        descendants = append(descendants, con)
    }
}

func createTask(con sway.Node, wsNum int64) task {
    t := task{}
    t.conID = con.ID
    if con.AppID != nil {
        t.ID = *con.AppID
    } else {
        wp := *con.WindowProperties
        t.ID = wp.Class
    }
    t.Name = con.Name
    t.PID = *con.PID
    t.WsNum = wsNum

    return t
}

func workspaceNum(workspaces []sway.Workspace, name string) int64 {
    for _, ws := range workspaces {
        if ws.Name == name {
            return ws.Num
        }
    }
    return 0
}

func pinnedButton(ID string) *gtk.Button {
    button, _ := gtk.ButtonNew()
    image, err := createImage(ID)
    if err == nil {
        button.SetImage(image)
        button.SetImagePosition(gtk.POS_TOP)
        button.SetAlwaysShowImage(true)
        button.SetLabel("")

        /*button.Connect("clicked", func() {
            onButtonClick(t.ID, t.conID)
        })*/

    } else {
        button.SetLabel(ID)
    }
    return button
}

func taskButton(t task, instances []task) *gtk.Button {
    fmt.Println("Creating button, conID = ", t.conID, t.ID)
    button, _ := gtk.ButtonNew()
    image, err := createImage(t.ID)
    if err == nil {
        button.SetImage(image)
        button.SetImagePosition(gtk.POS_TOP)
        button.SetAlwaysShowImage(true)
        // TODO: instead of the label, we'll need some graphics later
        button.SetLabel(strconv.Itoa(len(instances)))

        if len(instances) == 1 {
            button.Connect("clicked", func() {
                focusCon(t.conID)
            })
        }

    } else {
        button.SetLabel(t.ID)
    }
    return button
}

func taskMenu(taskID string, instances []task) gtk.Menu {
    menu, _ := gtk.MenuNew()

    iconName, _ := getIcon(taskID)
    for _, instance := range instances {
        menuItem, _ := gtk.MenuItemNew()
        hbox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
        image, _ := createImage(iconName)
        hbox.PackStart(image, false, false, 0)
        label, _ := gtk.LabelNew(fmt.Sprintf("%s (%v)", instance.Name, instance.WsNum))
        hbox.PackStart(label, false, false, 0)
        menuItem.Add(hbox)
        fmt.Println("MenuItem", iconName, instance.ID, instance.Name, instance.conID, instance.WsNum)
    }
    return *menu
}

func inPinned(taskID string) bool {
    for _, id := range pinned {
        if id == taskID {
            return true
        }
    }
    return false
}

func createImage(iconName string) (*gtk.Image, error) {
    pixbuf, err := createPixbuf(iconName, 48)
    if err != nil {
        return nil, err
    }
    image, _ := gtk.ImageNewFromPixbuf(pixbuf)

    return image, nil
}

func createPixbuf(icon string, size int) (*gdk.Pixbuf, error) {
    if strings.HasPrefix(icon, "/") {
        pixbuf, err := gdk.PixbufNewFromFileAtSize(icon, size, size)
        if err != nil {
            fmt.Println("Error Pixbuf.new_from_file_at_size: ", err)
            return nil, err
        }
        return pixbuf, nil
    }

    iconTheme, err := gtk.IconThemeGetDefault()
    if err != nil {
        log.Fatal("Couldn't get default theme: ", err)
    }
    pixbuf, err := iconTheme.LoadIcon(icon, size, gtk.ICON_LOOKUP_FORCE_SIZE)
    if err != nil {
        ico, err := getIcon(icon)
        if err != nil {
            return nil, err
        }

        if strings.HasPrefix(ico, "/") {
            pixbuf, err := gdk.PixbufNewFromFileAtSize(ico, size, size)
            if err != nil {
                return nil, err
            }
            return pixbuf, nil
        }

        pixbuf, err := iconTheme.LoadIcon(ico, size, gtk.ICON_LOOKUP_FORCE_SIZE)
        if err != nil {
            return nil, err
        }
        return pixbuf, nil
    }
    return pixbuf, nil
}

func cacheDir() string {
    if os.Getenv("XDG_CACHE_HOME") != "" {
        return os.Getenv("XDG_CONFIG_HOME")
    }
    if os.Getenv("HOME") != "" && pathExists(filepath.Join(os.Getenv("HOME"), ".cache")) {
        p := filepath.Join(os.Getenv("HOME"), ".cache")
        return p
    }
    return ""
}

func configDir() string {
    if os.Getenv("XDG_CONFIG_HOME") != "" {
        return (fmt.Sprintf("%s/nwg-dock", os.Getenv("XDG_CONFIG_HOME")))
    }
    return (fmt.Sprintf("%s/.config/nwg-dock", os.Getenv("HOME")))
}

func createDir(dir string) {
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        err := os.MkdirAll(dir, os.ModePerm)
        if err == nil {
            fmt.Println("Creating dir:", dir)
        }
    }
}

func getAppDirs() []string {
    var dirs []string
    xdgDataDirs := ""

    home := os.Getenv("HOME")
    xdgDataHome := os.Getenv("XDG_DATA_HOME")
    if os.Getenv("XDG_DATA_DIRS") != "" {
        xdgDataDirs = os.Getenv("XDG_DATA_DIRS")
    } else {
        xdgDataDirs = "/usr/local/share/:/usr/share/"
    }
    if xdgDataHome != "" {
        dirs = append(dirs, filepath.Join(xdgDataHome, "applications"))
    } else if home != "" {
        dirs = append(dirs, filepath.Join(home, ".local/share/applications"))
    }
    for _, d := range strings.Split(xdgDataDirs, ":") {
        dirs = append(dirs, filepath.Join(d, "applications"))
    }
    flatpakDirs := []string{filepath.Join(home, ".local/share/flatpak/exports/share/applications"),
        "/var/lib/flatpak/exports/share/applications"}

    for _, d := range flatpakDirs {
        if !isIn(dirs, d) {
            dirs = append(dirs, d)
        }
    }
    return dirs
}

func isIn(slice []string, val string) bool {
    for _, item := range slice {
        if item == val {
            return true
        }
    }
    return false
}

func getIcon(appName string) (string, error) {
    if strings.HasPrefix(strings.ToUpper(appName), "GIMP") {
        return "gimp", nil
    }
    for _, d := range appDirs {
        path := filepath.Join(d, fmt.Sprintf("%s.desktop", appName))
        p := ""
        if pathExists(path) {
            p = path
        } else if pathExists(strings.ToLower(path)) {
            p = strings.ToLower(path)
        }
        if p != "" {
            lines, err := loadTextFile(p)
            if err != nil {
                return "", err
            }
            for _, line := range lines {
                if strings.HasPrefix(strings.ToUpper(line), "ICON") {
                    return strings.Split(line, "=")[1], nil
                }
            }
        }
    }
    return "", errors.New("Couldn't find the icon")
}

func getExec(appName string) (string, error) {
    if strings.HasPrefix(strings.ToUpper(appName), "GIMP") {
        appName = "gimp"
    }
    for _, d := range appDirs {
        path := filepath.Join(d, fmt.Sprintf("%s.desktop", appName))
        p := ""
        if pathExists(path) {
            p = path
        } else if pathExists(strings.ToLower(path)) {
            p = strings.ToLower(path)
        }
        if p != "" {
            lines, err := loadTextFile(p)
            if err != nil {
                return "", err
            }
            for _, line := range lines {
                if strings.HasPrefix(strings.ToUpper(line), "EXEC") {
                    l := line[5:]
                    cutAt := strings.Index(l, "%")
                    if cutAt != -1 {
                        l = l[:cutAt-1]
                    }
                    return l, nil
                }
            }
        }
    }
    return "", errors.New("Couldn't find the exec")
}

func pathExists(name string) bool {
    if _, err := os.Stat(name); err != nil {
        if os.IsNotExist(err) {
            return false
        }
    }
    return true
}

func loadTextFile(path string) ([]string, error) {
    bytes, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }
    lines := strings.Split(string(bytes), "\n")
    var output []string
    for _, line := range lines {
        line = strings.TrimSpace(line)
        output = append(output, line)
    }
    return output, nil
}

func onButtonClick(ID string, conID int64) {
    exec, err := getExec(ID)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(exec)

    cmd := fmt.Sprintf("[con_id=%v] focus", conID)
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    client, err := sway.New(ctx)
    if err != nil {
        log.Panic(err)
    }
    client.RunCommand(ctx, cmd)
}

func focusCon(conID int64) {
    cmd := fmt.Sprintf("[con_id=%v] focus", conID)
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    client, err := sway.New(ctx)
    if err != nil {
        log.Panic(err)
    }
    client.RunCommand(ctx, cmd)
}
