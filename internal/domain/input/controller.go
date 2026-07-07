package input

// CursorController defines the domain interface for controlling the mouse, scroll, and keyboard on the host system.
type CursorController interface {
	Move(dx, dy float64)
	Drag(dx, dy float64)
	MouseDown()
	MouseUp()
	LeftClick()
	RightClick()
	Scroll(dx, dy int)
	Zoom(direction string)
	PlayPause()
	NextTrack()
	PreviousTrack()
	VolumeUp()
	VolumeDown()
	Mute()
	TypeString(text string)
	PressKey(keyName string)
	IsTrusted() bool
}
