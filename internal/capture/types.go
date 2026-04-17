package capture

import "image"

type Monitor struct {
	Name    string
	Aliases []string
	X       int
	Y       int
	Width   int
	Height  int
}

type Window struct {
	ID     WindowID
	Name   string
	X      int
	Y      int
	Width  int
	Height int
}

type WindowID int64

type ScreenCapture interface {
	ListMonitors() ([]Monitor, error)
	ListWindows() ([]Window, error)
	CaptureScreen(monitor string) (image.Image, error)
	CaptureWindow(id WindowID) (image.Image, error)
	CaptureRegion(x, y, w, h int) (image.Image, error)
	CaptureAllScreens() (image.Image, error)
}
