# Wails Android bridge and WebView integration.
-keepattributes *Annotation*
-keepattributes Signature
-keepattributes InnerClasses
-keepattributes EnclosingMethod
-keepattributes Exceptions

-keepclasseswithmembernames class * {
    native <methods>;
}

-keep class com.wails.app.** { *; }

-keepclassmembers class * {
    @android.webkit.JavascriptInterface <methods>;
}

-keepclassmembers class fqcn.of.javascript.interface.for.webview {
    public *;
}

-keep class androidx.webkit.** { *; }
-keep class androidx.biometric.** { *; }
-keep class androidx.security.crypto.** { *; }

-dontwarn org.chromium.**
-dontwarn androidx.**
-dontwarn javax.annotation.**
