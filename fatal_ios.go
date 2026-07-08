//go:build ios

// SPDX-License-Identifier: MIT

package main

/*
#import <UIKit/UIKit.h>

static void showFatalErrorAlert(const char* title, const char* message) {
    NSString *nsTitle = [NSString stringWithUTF8String:title];
    NSString *nsMessage = [NSString stringWithUTF8String:message];
    dispatch_async(dispatch_get_main_queue(), ^{
        __block void (^presentAlert)(void) = ^{
            UIViewController *topVC = nil;
            for (UIWindowScene *scene in [UIApplication sharedApplication].connectedScenes) {
                for (UIWindow *window in scene.windows) {
                    if (window.isKeyWindow) {
                        topVC = window.rootViewController;
                        break;
                    }
                }
                if (topVC) break;
            }
            if (!topVC) {
                topVC = [UIApplication sharedApplication].keyWindow.rootViewController;
            }
            while (topVC.presentedViewController) {
                topVC = topVC.presentedViewController;
            }
            if (topVC) {
                UIAlertController *alert = [UIAlertController alertControllerWithTitle:nsTitle
                                                                               message:nsMessage
                                                                        preferredStyle:UIAlertControllerStyleAlert];
                [alert addAction:[UIAlertAction actionWithTitle:@"OK" style:UIAlertActionStyleDefault handler:nil]];
                [topVC presentViewController:alert animated:YES completion:nil];
            } else {
                // Retry in 0.5 seconds if key window/root view controller is not ready yet
                dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(0.5 * NSEC_PER_SEC)), dispatch_get_main_queue(), presentAlert);
            }
        };
        presentAlert();
    });
}
*/
import "C"

import (
	"log"
	"unsafe"
)

func handleFatalError(err error) {
	log.Printf("FATAL ERROR: %v", err)

	title := "Application Error"
	message := err.Error()

	cTitle := C.CString(title)
	cMessage := C.CString(message)
	defer C.free(unsafe.Pointer(cTitle))
	defer C.free(unsafe.Pointer(cMessage))

	C.showFatalErrorAlert(cTitle, cMessage)

	// Block the thread forever so the app does not exit and the user can see the alert dialog.
	select {}
}
