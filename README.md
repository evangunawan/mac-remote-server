# Mac Remote Server

A lightweight, native macOS helper and dynamic web application that turns any mobile browser (Safari, Chrome, Firefox) into a wireless trackpad, keyboard, and media controller for your Mac.

---

## 🚀 Features

*   **Tactile Trackpad**:
    *   Smooth cursor movements with customizable sensitivity.
    *   Single-tap for Left Click.
    *   Two-finger tap for Right Click (with ghost-click prevention).
    *   Two-finger dragging for scrolling.
    *   Pinch gestures for Zoom (In/Out).
*   **Tactile Media Controller**:
    *   Master Play/Pause toggler.
    *   Skip Forward (Next) and Skip Backward (Previous) track controls.
    *   Volume Up, Volume Down, and Mute buttons.
    *   Subtle spring-loaded micro-wiggles on button clicks.
*   **Keyboard Emulation**:
    *   Pulsing "Tap to Type" capsule trigger that opens your phone's native keyboard.
    *   Emulates character entry, Backspaces, and Enter keystrokes remotely.
*   **Navigation D-Pad**:
    *   Circular grid-based directional pad.
    *   Up, Down, Left, and Right buttons mimicking keyboard arrow presses.
    *   Central **OK** button mimicking the Enter/Return key.
*   **Obsidian UI Design**:
    *   Vibrant dark and light themes with state persistence (`localStorage`).
    *   Safe-area integrations to avoid clipping behind Safari toolbars or iOS Home bars.

---

## 🛠️ Requirements & Setup

*   **Operating System**: macOS (requires Quartz Event Services).
*   **Permissions**: Accessibility access must be granted to the terminal or wrapper running the server under *System Settings > Privacy & Security > Accessibility*.

### Running the App
Run the pre-compiled standalone binary:
```bash
./mac-remote-server start
```

For developer hot-reloading mode (reading web assets directly from the filesystem without recompilation):
```bash
./mac-remote-server start -dev
```

Once running, connect your phone/tablet to the same Wi-Fi network and scan/visit the IP address displayed on the server startup screen.

---

## ✍️ Authorship

> [!NOTE]
> This entire application, including the Cgo native macOS integrations, the Clean Architecture Go server, and the dynamic glassmorphic web UI, was authored and refined by **Gemini 3.5 Flash**.
