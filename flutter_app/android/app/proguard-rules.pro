# Flutter-specific ProGuard/R8 rules.

# androidx.window optional extensions are referenced transitively (by plugins
# such as package_info_plus / android_intent_plus) but are not present on all
# devices. R8 otherwise fails the release build with "Missing class".
-dontwarn androidx.window.extensions.**
-dontwarn androidx.window.sidecar.**
