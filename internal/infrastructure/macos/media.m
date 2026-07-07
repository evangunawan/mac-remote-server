// media.m - Native macOS media key simulation using NSEvent
// This replicates exactly what happens when you press a physical media key (Fn+F7/F8/F9).
#import <Cocoa/Cocoa.h>
#include "media.h"

void SimulateNativeMediaKey(int key) {
    @autoreleasepool {
        // Key Down event - modifierFlags 0xa00 indicates key down for NX_SYSDEFINED
        NSEvent *downEvent = [NSEvent otherEventWithType:NSEventTypeSystemDefined
                                                location:NSMakePoint(0, 0)
                                           modifierFlags:0xa00
                                               timestamp:0
                                            windowNumber:0
                                                 context:nil
                                                 subtype:8
                                                   data1:((key << 16) | 0x0a00)
                                                   data2:-1];
        CGEventPost(kCGHIDEventTap, [downEvent CGEvent]);

        // Key Up event - modifierFlags 0xb00 indicates key up for NX_SYSDEFINED
        NSEvent *upEvent = [NSEvent otherEventWithType:NSEventTypeSystemDefined
                                              location:NSMakePoint(0, 0)
                                         modifierFlags:0xb00
                                             timestamp:0
                                          windowNumber:0
                                               context:nil
                                               subtype:8
                                                 data1:((key << 16) | 0x0b00)
                                                 data2:-1];
        CGEventPost(kCGHIDEventTap, [upEvent CGEvent]);
    }
}
