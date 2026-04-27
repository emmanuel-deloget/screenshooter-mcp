package wayland

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/nskaggs/perfuncted/window"
)

const (
	gnomeExtBus   = "org.screenshooter.mcp"
	gnomeExtPath  = "/org/screenshooter/mcp"
	gnomeExtIface = "org.screenshooter.mcp.Windows"
)

type GnomeManager struct {
	conn *dbus.Conn
	obj  dbus.BusObject
}

func NewGnomeManager() (*GnomeManager, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("gnome: session bus: %w", err)
	}
	obj := conn.Object(gnomeExtBus, gnomeExtPath)
	// Probe
	var windows []map[string]dbus.Variant
	if err := obj.Call(gnomeExtIface+".List", 0).Store(&windows); err != nil {
		return nil, fmt.Errorf("gnome: extension not available: %w", err)
	}
	return &GnomeManager{conn: conn, obj: obj}, nil
}

func (g *GnomeManager) List(ctx context.Context) ([]window.Info, error) {
	var raw []map[string]dbus.Variant
	if err := g.obj.CallWithContext(ctx, gnomeExtIface+".List", 0).Store(&raw); err != nil {
		return nil, fmt.Errorf("gnome: list: %w", err)
	}
	out := make([]window.Info, 0, len(raw))
	for _, e := range raw {
		out = append(out, window.Info{
			ID:        e["id"].Value().(uint64),
			Title:     e["title"].Value().(string),
			PID:       e["pid"].Value().(int32),
			X:         int(e["x"].Value().(int32)),
			Y:         int(e["y"].Value().(int32)),
			W:         int(e["w"].Value().(int32)),
			H:         int(e["h"].Value().(int32)),
			Minimized: e["minimized"].Value().(bool),
			Maximized: e["maximized"].Value().(bool),
		})
	}
	return out, nil
}

func (g *GnomeManager) Activate(ctx context.Context, title string) error {
	return g.obj.CallWithContext(ctx, gnomeExtIface+".Activate", 0, title).Err
}

func (g *GnomeManager) Move(ctx context.Context, title string, x, y int) error {
	return g.obj.CallWithContext(ctx, gnomeExtIface+".Move", 0, title, int32(x), int32(y)).Err
}

func (g *GnomeManager) Resize(ctx context.Context, title string, w, h int) error {
	return g.obj.CallWithContext(ctx, gnomeExtIface+".Resize", 0, title, int32(w), int32(h)).Err
}

func (g *GnomeManager) ActiveTitle(ctx context.Context) (string, error) {
	var title string
	if err := g.obj.CallWithContext(ctx, gnomeExtIface+".ActiveTitle", 0).Store(&title); err != nil {
		return "", fmt.Errorf("gnome: active title: %w", err)
	}
	return title, nil
}

func (g *GnomeManager) CloseWindow(ctx context.Context, title string) error {
	return g.obj.CallWithContext(ctx, gnomeExtIface+".Close", 0, title).Err
}

func (g *GnomeManager) Minimize(ctx context.Context, title string) error {
	return g.obj.CallWithContext(ctx, gnomeExtIface+".Minimize", 0, title).Err
}

func (g *GnomeManager) Maximize(ctx context.Context, title string) error {
	return g.obj.CallWithContext(ctx, gnomeExtIface+".Maximize", 0, title).Err
}

func (g *GnomeManager) Restore(ctx context.Context, title string) error {
	return g.Maximize(ctx, title)
}

func (g *GnomeManager) Close() error {
	return g.conn.Close()
}
