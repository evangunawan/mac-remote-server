// Enable active CSS states on iOS Safari
document.body.addEventListener('touchstart', () => {}, { passive: true });

// WebSocket state and elements
let ws;
const statusEl = document.getElementById('connection-status');
const statusTextEl = statusEl.querySelector('.status-text');

const trackpad = document.getElementById('trackpad');
const scrollZone = document.getElementById('scroll-zone');
const btnLeftClick = document.getElementById('btn-left-click');
const btnRightClick = document.getElementById('btn-right-click');
const sliderSensitivity = document.getElementById('slider-sensitivity');
const valSensitivity = document.getElementById('val-sensitivity');

// Settings Modal elements
const btnSettings = document.getElementById('btn-settings');
const btnCloseSettings = document.getElementById('btn-close-settings');
const modalSettings = document.getElementById('modal-settings');

// Theme Switcher elements
const btnThemeDark = document.getElementById('btn-theme-dark');
const btnThemeLight = document.getElementById('btn-theme-light');

let sensitivity = parseFloat(sliderSensitivity.value);

// Touch tracking variables
let isTouching = false;
let startX = 0;
let startY = 0;
let lastX = 0;
let lastY = 0;
let touchStartTime = 0;

// Two-finger tap for right-click variables
let isTwoFingerTouching = false;
let twoFingerTouchStartTime = 0;
let twoFingerStartX = 0;
let twoFingerStartY = 0;
let preventOneFingerClick = false;

// Scrolling variables
let lastScrollX = 0;
let lastScrollY = 0;
let lastScrollZoneY = 0;
let lastPinchDist = 0;

// Scroll sensitivity multiplier (lower = slower scrolling)
const SCROLL_SENSITIVITY = 0.4;

// Initialize Theme from localStorage
function initTheme() {
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme === 'light') {
        document.body.classList.add('light-mode');
        btnThemeLight.classList.add('active');
        btnThemeDark.classList.remove('active');
    } else {
        document.body.classList.remove('light-mode');
        btnThemeDark.classList.add('active');
        btnThemeLight.classList.remove('active');
    }
}

// Theme toggles
btnThemeDark.addEventListener('click', () => {
    document.body.classList.remove('light-mode');
    btnThemeDark.classList.add('active');
    btnThemeLight.classList.remove('active');
    localStorage.setItem('theme', 'dark');
});

btnThemeLight.addEventListener('click', () => {
    document.body.classList.add('light-mode');
    btnThemeLight.classList.add('active');
    btnThemeDark.classList.remove('active');
    localStorage.setItem('theme', 'light');
});

// Sensitivity changes
sliderSensitivity.addEventListener('input', (e) => {
    sensitivity = parseFloat(e.target.value);
    valSensitivity.textContent = sensitivity.toFixed(1) + 'x';
});

// Settings Modal triggers
btnSettings.addEventListener('click', () => {
    modalSettings.classList.add('open');
});

btnCloseSettings.addEventListener('click', () => {
    modalSettings.classList.remove('open');
});

// Close modal when tapping outside of it
modalSettings.addEventListener('click', (e) => {
    if (e.target === modalSettings) {
        modalSettings.classList.remove('open');
    }
});

// Establish WebSocket Connection
function connect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    
    statusTextEl.textContent = 'Connecting...';
    statusEl.className = 'status disconnected';
    
    ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
        statusTextEl.textContent = 'Connected';
        statusEl.className = 'status connected';
        console.log('Connected to Mac Remote Server');
    };
    
    ws.onclose = () => {
        statusTextEl.textContent = 'Disconnected';
        statusEl.className = 'status disconnected';
        console.log('Connection lost. Reconnecting in 3s...');
        setTimeout(connect, 3000);
    };
    
    ws.onerror = (err) => {
        console.error('WebSocket Error:', err);
    };
}

// WS Helper
function send(msg) {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify(msg));
    }
}

// Click buttons
btnLeftClick.addEventListener('click', () => {
    send({ action: 'click', button: 'left' });
});

btnRightClick.addEventListener('click', () => {
    send({ action: 'click', button: 'right' });
});

// Touch event handlers for trackpad
trackpad.addEventListener('touchstart', (e) => {
    if (e.touches.length === 1) {
        if (preventOneFingerClick) {
            return;
        }
        isTouching = true;
        touchStartTime = Date.now();
        lastX = e.touches[0].clientX;
        lastY = e.touches[0].clientY;
        startX = lastX;
        startY = lastY;
        isTwoFingerTouching = false;
    } else if (e.touches.length === 2) {
        isTouching = false; // Cancel one finger move
        isTwoFingerTouching = true;
        twoFingerTouchStartTime = Date.now();
        twoFingerStartX = (e.touches[0].clientX + e.touches[1].clientX) / 2;
        twoFingerStartY = (e.touches[0].clientY + e.touches[1].clientY) / 2;

        const dx = e.touches[0].clientX - e.touches[1].clientX;
        const dy = e.touches[0].clientY - e.touches[1].clientY;
        lastPinchDist = Math.sqrt(dx * dx + dy * dy);

        lastScrollX = (e.touches[0].clientX + e.touches[1].clientX) / 2;
        lastScrollY = (e.touches[0].clientY + e.touches[1].clientY) / 2;
    }
}, { passive: false });

trackpad.addEventListener('touchmove', (e) => {
    e.preventDefault(); // Prevent iOS page scrolling
    
    if (e.touches.length === 1 && isTouching) {
        const x = e.touches[0].clientX;
        const y = e.touches[0].clientY;
        
        const dx = (x - lastX) * sensitivity;
        const dy = (y - lastY) * sensitivity;
        
        send({ action: 'move', dx: dx, dy: dy });
        
        lastX = x;
        lastY = y;
    } else if (e.touches.length === 2) {
        const touch1 = e.touches[0];
        const touch2 = e.touches[1];
        
        // Calculate current distance for pinch detection
        const pDx = touch1.clientX - touch2.clientX;
        const pDy = touch1.clientY - touch2.clientY;
        const dist = Math.sqrt(pDx * pDx + pDy * pDy);
        
        // Calculate center point for scroll detection
        const x = (touch1.clientX + touch2.clientX) / 2;
        const y = (touch1.clientY + touch2.clientY) / 2;
        
        // If they move too much, cancel the right-click two-finger tap detection.
        // Threshold is generous: two-finger taps naturally wobble as the fingers
        // land/lift out of sync, and a tight bound would cancel almost every tap.
        const moveDist = Math.sqrt(Math.pow(x - twoFingerStartX, 2) + Math.pow(y - twoFingerStartY, 2));
        if (moveDist > 20) {
            isTwoFingerTouching = false;
        }

        const distChange = dist - lastPinchDist;
        const scrollDx = x - lastScrollX;
        const scrollDy = y - lastScrollY;
        
        // If absolute distance change is larger than a panning threshold, zoom!
        if (Math.abs(distChange) > 8) {
            if (distChange > 0) {
                send({ action: 'zoom', direction: 'in' });
            } else {
                send({ action: 'zoom', direction: 'out' });
            }
            lastPinchDist = dist;
        } else if (Math.abs(scrollDx) > 1 || Math.abs(scrollDy) > 1) {
            // Scroll action
            send({ action: 'scroll', dx: -Math.round(scrollDx * SCROLL_SENSITIVITY), dy: -Math.round(scrollDy * SCROLL_SENSITIVITY) });
        }
        
        lastScrollX = x;
        lastScrollY = y;
        lastPinchDist = dist;
    }
}, { passive: false });

trackpad.addEventListener('touchend', (e) => {
    if (isTwoFingerTouching) {
        isTwoFingerTouching = false;
        const duration = Date.now() - twoFingerTouchStartTime;

        if (duration < 400) {
            send({ action: 'click', button: 'right' });
            preventOneFingerClick = true;
            setTimeout(() => {
                preventOneFingerClick = false;
            }, 300); // 300ms cooldown to block trailing lift touch
        }
    }

    if (isTouching) {
        isTouching = false;
        if (preventOneFingerClick) {
            return;
        }
        const duration = Date.now() - touchStartTime;
        const endX = e.changedTouches[0].clientX;
        const endY = e.changedTouches[0].clientY;
        const dist = Math.sqrt(Math.pow(endX - startX, 2) + Math.pow(endY - startY, 2));
        
        // Short tap with minimal displacement triggers left click
        if (duration < 250 && dist < 8) {
            send({ action: 'click', button: 'left' });
        }
    }
}, { passive: false });

// Touch event handlers for dedicated scroll strip
scrollZone.addEventListener('touchstart', (e) => {
    if (e.touches.length === 1) {
        lastScrollZoneY = e.touches[0].clientY;
    }
}, { passive: false });

scrollZone.addEventListener('touchmove', (e) => {
    e.preventDefault();
    if (e.touches.length === 1) {
        const y = e.touches[0].clientY;
        // Strip is a dedicated scroller, so keep it a touch faster than the trackpad.
        const dy = (y - lastScrollZoneY) * (SCROLL_SENSITIVITY * 2);
        
        // Vertical scroll only. Invert direction for natural scroll behavior.
        send({ action: 'scroll', dx: 0, dy: -Math.round(dy) });
        
        lastScrollZoneY = y;
    }
}, { passive: false });

// Tab switching logic
const tabs = document.querySelectorAll('.tab-btn');
const panels = document.querySelectorAll('.panel');

tabs.forEach(tab => {
    tab.addEventListener('click', () => {
        tabs.forEach(t => t.classList.remove('active'));
        panels.forEach(p => p.classList.remove('active'));
        
        tab.classList.add('active');
        const targetPanel = document.getElementById(tab.getAttribute('data-target'));
        if (targetPanel) {
            targetPanel.classList.add('active');
        }
    });
});

// Media controls logic
document.getElementById('media-prev').addEventListener('click', () => send({ action: 'previous' }));
document.getElementById('media-play').addEventListener('click', () => send({ action: 'playpause' }));
document.getElementById('media-next').addEventListener('click', () => send({ action: 'next' }));
document.getElementById('media-voldown').addEventListener('click', () => send({ action: 'voldown' }));
document.getElementById('media-mute').addEventListener('click', () => send({ action: 'mute' }));
document.getElementById('media-volup').addEventListener('click', () => send({ action: 'volup' }));

// Navigation D-Pad controls logic
document.getElementById('nav-up').addEventListener('click', () => send({ action: 'key', name: 'up' }));
document.getElementById('nav-down').addEventListener('click', () => send({ action: 'key', name: 'down' }));
document.getElementById('nav-left').addEventListener('click', () => send({ action: 'key', name: 'left' }));
document.getElementById('nav-right').addEventListener('click', () => send({ action: 'key', name: 'right' }));
document.getElementById('nav-enter').addEventListener('click', () => send({ action: 'key', name: 'enter' }));

// Keyboard and remote typing logic
const keyboardTrigger = document.getElementById('keyboard-trigger');
const hiddenKeyboard = document.getElementById('hidden-keyboard');

keyboardTrigger.addEventListener('touchstart', (e) => {
    e.preventDefault();
    hiddenKeyboard.focus();
});

hiddenKeyboard.addEventListener('input', (e) => {
    const val = e.target.value;
    if (val) {
        send({ action: 'type', text: val });
        e.target.value = ''; // Reset immediately
    }
});

hiddenKeyboard.addEventListener('keydown', (e) => {
    if (e.key === 'Backspace') {
        send({ action: 'key', name: 'backspace' });
    } else if (e.key === 'Enter') {
        send({ action: 'key', name: 'enter' });
    }
});

// Init theme state
initTheme();

// Connect WebSocket
connect();
