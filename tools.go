package main

import (
    "context"
    "fmt"
    "log"
    "sort"
    "strings"
    "time"

    "github.com/gotk3/gotk3/gdk"
    "github.com/gotk3/gotk3/gtk"
    "github.com/joshuarubin/go-sway"
)

var descendants []sway.Node

type task struct {
    ID    string
    Name  string
    PID   uint32
    WsNum int64
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

func createButton(iconName string, wsNum int64) *gtk.Button {
    button, _ := gtk.ButtonNew()
    image, err := createImage(iconName)
    if err == nil {
        button.SetImage(image)
        button.SetImagePosition(gtk.POS_TOP)
        button.SetAlwaysShowImage(true)
        button.SetLabel(fmt.Sprintf("%2d", wsNum))
    } else {
        button.SetLabel(iconName)
    }
    (*button).SetSizeRequest(60, 60)
    return button
}

func createImage(iconName string) (*gtk.Image, error) {
    pixbuf, err := createPixbuf(iconName, 30)
    if err != nil {
        return nil, err
    }
    image, _ := gtk.ImageNewFromPixbuf(pixbuf)

    return image, nil
}

func createPixbuf(icon string, size int) (*gdk.Pixbuf, error) {
    iconTheme, err := gtk.IconThemeGetDefault()
    if err != nil {
        log.Fatal("Couldn't get default theme: ", err)
    }
    pixbuf, err := iconTheme.LoadIcon(icon, size, gtk.ICON_LOOKUP_FORCE_SIZE)
    if err != nil {
        return nil, err
    }
    return pixbuf, nil
}
