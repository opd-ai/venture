# Add project specific ProGuard rules here.
# By default, the flags in this file are appended to flags specified
# in ${ANDROID_HOME}/tools/proguard/proguard-android.txt

# Keep all Ebiten/Go mobile classes
-keep class org.ebitengine.gomobile.** { *; }
-keep class go.** { *; }
-keep class mobile.** { *; }

# Keep native methods
-keepclasseswithmembernames class * {
    native <methods>;
}

# Keep MainActivity and EbitenView
-keep class com.venture.game.MainActivity { *; }
-keep class com.venture.game.mobile.EbitenView { *; }

# Suppress warnings for Go generated code
-dontwarn go.**
-dontwarn mobile.**
-dontwarn org.ebitengine.gomobile.**
