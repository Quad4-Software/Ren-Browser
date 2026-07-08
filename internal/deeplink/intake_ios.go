//go:build ios

package deeplink

import "C"

//export RenBrowserHandleOpenURL
func RenBrowserHandleOpenURL(urlCString *C.char) {
	if urlCString == nil {
		return
	}
	_, _ = HandleIncoming(C.GoString(urlCString))
}
