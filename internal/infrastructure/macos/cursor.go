package macos

/*
#cgo LDFLAGS: -framework ApplicationServices -framework CoreGraphics
#include <ApplicationServices/ApplicationServices.h>
#include <CoreGraphics/CoreGraphics.h>
#include <unistd.h>

void MoveMouse(double dx, double dy) {
    CGEventRef event = CGEventCreate(NULL);
    if (!event) return;
    CGPoint current = CGEventGetLocation(event);
    CFRelease(event);

    CGPoint newPos = CGPointMake(current.x + dx, current.y + dy);

    CGEventRef moveEvent = CGEventCreateMouseEvent(NULL, kCGEventMouseMoved, newPos, kCGMouseButtonLeft);
    if (moveEvent) {
        CGEventPost(kCGHIDEventTap, moveEvent);
        CFRelease(moveEvent);
    }
}

void DragMouse(double dx, double dy) {
    CGEventRef event = CGEventCreate(NULL);
    if (!event) return;
    CGPoint current = CGEventGetLocation(event);
    CFRelease(event);

    CGPoint newPos = CGPointMake(current.x + dx, current.y + dy);

    CGEventRef dragEvent = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseDragged, newPos, kCGMouseButtonLeft);
    if (dragEvent) {
        CGEventPost(kCGHIDEventTap, dragEvent);
        CFRelease(dragEvent);
    }
}

void MouseDown() {
    CGEventRef event = CGEventCreate(NULL);
    if (!event) return;
    CGPoint current = CGEventGetLocation(event);
    CFRelease(event);

    CGEventRef down = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseDown, current, kCGMouseButtonLeft);
    if (down) {
        CGEventSetIntegerValueField(down, kCGMouseEventClickState, 1);
        CGEventPost(kCGHIDEventTap, down);
        CFRelease(down);
    }
}

void MouseUp() {
    CGEventRef event = CGEventCreate(NULL);
    if (!event) return;
    CGPoint current = CGEventGetLocation(event);
    CFRelease(event);

    CGEventRef up = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseUp, current, kCGMouseButtonLeft);
    if (up) {
        CGEventSetIntegerValueField(up, kCGMouseEventClickState, 1);
        CGEventPost(kCGHIDEventTap, up);
        CFRelease(up);
    }
}

void LeftClick() {
    CGEventRef event = CGEventCreate(NULL);
    if (!event) return;
    CGPoint current = CGEventGetLocation(event);
    CFRelease(event);

    CGEventRef down = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseDown, current, kCGMouseButtonLeft);
    CGEventRef up = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseUp, current, kCGMouseButtonLeft);

    if (down && up) {
        CGEventSetIntegerValueField(down, kCGMouseEventClickState, 1);
        CGEventSetIntegerValueField(up, kCGMouseEventClickState, 1);
        CGEventPost(kCGHIDEventTap, down);
        usleep(10000); // 10ms
        CGEventPost(kCGHIDEventTap, up);
    }
    if (down) CFRelease(down);
    if (up) CFRelease(up);
}

void RightClick() {
    CGEventRef event = CGEventCreate(NULL);
    if (!event) return;
    CGPoint current = CGEventGetLocation(event);
    CFRelease(event);

    CGEventRef down = CGEventCreateMouseEvent(NULL, kCGEventRightMouseDown, current, kCGMouseButtonRight);
    CGEventRef up = CGEventCreateMouseEvent(NULL, kCGEventRightMouseUp, current, kCGMouseButtonRight);

    if (down && up) {
        CGEventSetIntegerValueField(down, kCGMouseEventClickState, 1);
        CGEventSetIntegerValueField(up, kCGMouseEventClickState, 1);
        CGEventPost(kCGHIDEventTap, down);
        usleep(10000); // 10ms
        CGEventPost(kCGHIDEventTap, up);
    }
    if (down) CFRelease(down);
    if (up) CFRelease(up);
}

void ScrollMouse(int dy, int dx) {
    CGEventRef scrollEvent = CGEventCreateScrollWheelEvent(NULL, kCGScrollEventUnitPixel, 2, dy, dx);
    if (scrollEvent) {
        CGEventPost(kCGHIDEventTap, scrollEvent);
        CFRelease(scrollEvent);
    }
}

void GetMousePos(double *x, double *y) {
    CGEventRef event = CGEventCreate(NULL);
    if (event) {
        CGPoint current = CGEventGetLocation(event);
        *x = current.x;
        *y = current.y;
        CFRelease(event);
    }
}

void ZoomMouse(int zoomIn) {
    CGKeyCode keyCode = zoomIn ? (CGKeyCode)24 : (CGKeyCode)27; // 24 = '=', 27 = '-'
    CGEventRef down = CGEventCreateKeyboardEvent(NULL, keyCode, true);
    CGEventRef up = CGEventCreateKeyboardEvent(NULL, keyCode, false);
    if (down && up) {
        CGEventSetFlags(down, kCGEventFlagMaskCommand);
        CGEventSetFlags(up, kCGEventFlagMaskCommand);
        CGEventPost(kCGHIDEventTap, down);
        usleep(10000); // 10ms
        CGEventPost(kCGHIDEventTap, up);
    }
    if (down) CFRelease(down);
    if (up) CFRelease(up);
}

#define NX_SYSDEFINED 14
#define NX_SUBTYPE_AUX_CONTROL_BUTTONS 8

static const CGEventField kCGEventSubtype = (CGEventField)8;
static const CGEventField kCGEventData1 = (CGEventField)149;
static const CGEventField kCGEventData2 = (CGEventField)150;

void SimulateMediaKey(int key) {
    // Key Down event
    CGEventRef keyDown = CGEventCreate(NULL);
    if (keyDown) {
        CGEventSetType(keyDown, NX_SYSDEFINED);
        CGEventSetIntegerValueField(keyDown, kCGEventSubtype, NX_SUBTYPE_AUX_CONTROL_BUTTONS);
        CGEventSetIntegerValueField(keyDown, kCGEventData1, (key << 16) | 0xa00); // 0xa00 = down flag
        CGEventSetIntegerValueField(keyDown, kCGEventData2, -1);
        CGEventPost(kCGHIDEventTap, keyDown);
        CFRelease(keyDown);
    }

    // Key Up event
    CGEventRef keyUp = CGEventCreate(NULL);
    if (keyUp) {
        CGEventSetType(keyUp, NX_SYSDEFINED);
        CGEventSetIntegerValueField(keyUp, kCGEventSubtype, NX_SUBTYPE_AUX_CONTROL_BUTTONS);
        CGEventSetIntegerValueField(keyUp, kCGEventData1, (key << 16) | 0xb00); // 0xb00 = up flag
        CGEventSetIntegerValueField(keyUp, kCGEventData2, -1);
        CGEventPost(kCGHIDEventTap, keyUp);
        CFRelease(keyUp);
    }
}

void TypeString(const char* text) {
    CGEventRef ev = CGEventCreateKeyboardEvent(NULL, 0, true);
    if (!ev) return;
    
    CFStringRef str = CFStringCreateWithCString(NULL, text, kCFStringEncodingUTF8);
    if (str) {
        CFIndex length = CFStringGetLength(str);
        UniChar *buffer = malloc(length * sizeof(UniChar));
        if (buffer) {
            CFStringGetCharacters(str, CFRangeMake(0, length), buffer);
            CGEventKeyboardSetUnicodeString(ev, length, buffer);
            CGEventPost(kCGHIDEventTap, ev);
            free(buffer);
        }
        CFRelease(str);
    }
    CFRelease(ev);
}

void PressSpecialKey(const char* keyName) {
    CGKeyCode keyCode = 0;
    if (strcmp(keyName, "backspace") == 0) {
        keyCode = 51;
    } else if (strcmp(keyName, "enter") == 0) {
        keyCode = 36;
    } else if (strcmp(keyName, "space") == 0) {
        keyCode = 49;
    } else if (strcmp(keyName, "left") == 0) {
        keyCode = 123;
    } else if (strcmp(keyName, "right") == 0) {
        keyCode = 124;
    } else if (strcmp(keyName, "down") == 0) {
        keyCode = 125;
    } else if (strcmp(keyName, "up") == 0) {
        keyCode = 126;
    } else {
        return;
    }
    
    CGEventRef down = CGEventCreateKeyboardEvent(NULL, keyCode, true);
    CGEventRef up = CGEventCreateKeyboardEvent(NULL, keyCode, false);
    if (down && up) {
        CGEventPost(kCGHIDEventTap, down);
        CGEventPost(kCGHIDEventTap, up);
    }
    if (down) CFRelease(down);
    if (up) CFRelease(up);
}

// IsTextFieldFocused removed

int IsProcessTrusted() {
    return AXIsProcessTrusted() ? 1 : 0;
}
*/
import "C"
import (
	"mac-remote-server/internal/domain/input"
	"unsafe"
)

type MacCursorController struct{}

// NewMacCursorController instantiates a new macOS-native CursorController.
func NewMacCursorController() input.CursorController {
	return &MacCursorController{}
}

func (m *MacCursorController) Move(dx, dy float64) {
	C.MoveMouse(C.double(dx), C.double(dy))
}

func (m *MacCursorController) Drag(dx, dy float64) {
	C.DragMouse(C.double(dx), C.double(dy))
}

func (m *MacCursorController) MouseDown() {
	C.MouseDown()
}

func (m *MacCursorController) MouseUp() {
	C.MouseUp()
}

func (m *MacCursorController) LeftClick() {
	C.LeftClick()
}

func (m *MacCursorController) RightClick() {
	C.RightClick()
}

func (m *MacCursorController) Scroll(dx, dy int) {
	// C.ScrollMouse takes (dy, dx) matching CGEventCreateScrollWheelEvent argument order (vertical wheel1, horizontal wheel2)
	C.ScrollMouse(C.int(dy), C.int(dx))
}

func (m *MacCursorController) Zoom(direction string) {
	if direction == "in" {
		C.ZoomMouse(1)
	} else if direction == "out" {
		C.ZoomMouse(0)
	}
}

func (m *MacCursorController) PlayPause() {
	C.SimulateMediaKey(16)
}

func (m *MacCursorController) NextTrack() {
	C.SimulateMediaKey(19)
}

func (m *MacCursorController) PreviousTrack() {
	C.SimulateMediaKey(20)
}

func (m *MacCursorController) VolumeUp() {
	C.SimulateMediaKey(0)
}

func (m *MacCursorController) VolumeDown() {
	C.SimulateMediaKey(1)
}

func (m *MacCursorController) Mute() {
	C.SimulateMediaKey(7)
}

func (m *MacCursorController) TypeString(text string) {
	cStr := C.CString(text)
	defer C.free(unsafe.Pointer(cStr))
	C.TypeString(cStr)
}

func (m *MacCursorController) PressKey(keyName string) {
	cStr := C.CString(keyName)
	defer C.free(unsafe.Pointer(cStr))
	C.PressSpecialKey(cStr)
}

// IsTextInputFocused removed

func (m *MacCursorController) IsTrusted() bool {
	return C.IsProcessTrusted() != 0
}

// GetMousePosition is a helper to retrieve current screen coordinates of the cursor (useful for tests).
func GetMousePosition() (x, y float64) {
	var cx, cy C.double
	C.GetMousePos(&cx, &cy)
	return float64(cx), float64(cy)
}
