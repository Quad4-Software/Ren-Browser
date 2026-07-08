//go:build ios
#import <UIKit/UIKit.h>

extern void RenBrowserHandleOpenURL(char *url);

@interface WailsAppDelegate : UIResponder <UIApplicationDelegate>
@end

@interface WailsAppDelegate (RenBrowserDeepLink)
@end

@implementation WailsAppDelegate (RenBrowserDeepLink)

- (BOOL)application:(UIApplication *)app openURL:(NSURL *)url options:(NSDictionary<UIApplicationOpenURLOptionsKey,id> *)options {
    if (url != nil) {
        RenBrowserHandleOpenURL((char *)[[url absoluteString] UTF8String]);
    }
    return YES;
}

- (BOOL)application:(UIApplication *)application continueUserActivity:(NSUserActivity *)userActivity restorationHandler:(void (^)(NSArray<id<UIUserActivityRestoring>> * _Nullable))restorationHandler {
    if ([userActivity.activityType isEqualToString:NSUserActivityTypeBrowsingWeb]) {
        NSURL *url = userActivity.webpageURL;
        if (url != nil) {
            RenBrowserHandleOpenURL((char *)[[url absoluteString] UTF8String]);
            return YES;
        }
    }
    return NO;
}

@end
