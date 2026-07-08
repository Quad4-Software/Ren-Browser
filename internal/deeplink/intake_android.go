//go:build android && cgo

package deeplink

/*
#include <jni.h>
#include <stdlib.h>

static const char* jstringToC(JNIEnv *env, jstring jstr) {
    if (jstr == NULL) return NULL;
    return (*env)->GetStringUTFChars(env, jstr, NULL);
}

static void releaseJString(JNIEnv *env, jstring jstr, const char* cstr) {
    if (jstr != NULL && cstr != NULL) {
        (*env)->ReleaseStringUTFChars(env, jstr, cstr);
    }
}
*/
import "C"

//export Java_com_wails_app_WailsBridge_nativeHandleDeepLink
func Java_com_wails_app_WailsBridge_nativeHandleDeepLink(env *C.JNIEnv, obj C.jobject, jurl C.jstring) {
	cURL := C.jstringToC(env, jurl)
	raw := C.GoString(cURL)
	C.releaseJString(env, jurl, cURL)
	_, _ = HandleIncoming(raw)
}
