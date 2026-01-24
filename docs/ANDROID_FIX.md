# Android ClassNotFoundException Fix

## Problem

The Android app was crashing on launch with:
```
java.lang.ClassNotFoundException: Didn't find class "org.ebitengine.gomobile.GoNativeActivity"
```

## Root Cause

The `AndroidManifest.xml` was configured to use `org.ebitengine.gomobile.GoNativeActivity` as the main activity class. However, with Ebiten v2.9+, the `ebitenmobile bind` tool generates an AAR containing `EbitenView` (a custom Android View), not a pre-built activity class.

The older GoNativeActivity approach was used in earlier versions of gomobile/ebitengine, but the modern approach requires developers to create their own Activity that integrates the EbitenView.

## Solution

Created a custom `MainActivity.java` that properly integrates with the ebitenmobile-generated AAR:

### 1. MainActivity.java (`build/android/src/main/java/com/venture/game/MainActivity.java`)

```java
package com.venture.game;

import android.app.Activity;
import android.os.Bundle;
import android.view.View;
import android.view.WindowManager;

import go.Seq;
import mobile.EbitenView;

public class MainActivity extends Activity {
    private EbitenView ebitenView;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        
        // Initialize Go mobile runtime
        Seq.setContext(getApplicationContext());
        
        // Keep screen on during gameplay
        getWindow().addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON);
        
        // Hide system UI for immersive fullscreen experience
        hideSystemUI();
        
        // Create and set the EbitenView
        ebitenView = new EbitenView(this);
        setContentView(ebitenView);
        
        // Start the game
        mobile.Mobile.start();
    }

    @Override
    protected void onPause() {
        super.onPause();
        if (ebitenView != null) {
            ebitenView.onPause();
        }
    }

    @Override
    protected void onResume() {
        super.onResume();
        if (ebitenView != null) {
            ebitenView.onResume();
        }
        hideSystemUI();
    }

    @Override
    public void onWindowFocusChanged(boolean hasFocus) {
        super.onWindowFocusChanged(hasFocus);
        if (hasFocus) {
            hideSystemUI();
        }
    }

    private void hideSystemUI() {
        View decorView = getWindow().getDecorView();
        decorView.setSystemUiVisibility(
                View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY
                | View.SYSTEM_UI_FLAG_LAYOUT_STABLE
                | View.SYSTEM_UI_FLAG_LAYOUT_HIDE_NAVIGATION
                | View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN
                | View.SYSTEM_UI_FLAG_HIDE_NAVIGATION
                | View.SYSTEM_UI_FLAG_FULLSCREEN);
    }
}
```

**Key responsibilities:**
- Initialize Go mobile runtime with `Seq.setContext()`
- Create and display the `EbitenView` 
- Call `mobile.Mobile.start()` to initialize the game
- Handle activity lifecycle (pause/resume)
- Provide immersive fullscreen experience

### 2. AndroidManifest.xml Updates

Changed from:
```xml
<activity
    android:name="org.ebitengine.gomobile.GoNativeActivity"
    ...>
    <meta-data android:name="android.app.lib_name" android:value="mobile" />
    ...
</activity>
```

To:
```xml
<activity
    android:name="com.venture.game.MainActivity"
    ...>
    ...
</activity>
```

**Changes:**
- Removed `org.ebitengine.gomobile.GoNativeActivity` reference
- Changed to `com.venture.game.MainActivity`
- Removed `meta-data` tag (not needed with custom activity)

### 3. build.gradle Updates

Added Java source directory to sourceSets:
```gradle
sourceSets {
    main {
        manifest.srcFile 'AndroidManifest.xml'
        java.srcDirs = ['src/main/java']  // Added this line
        res.srcDirs = ['res', 'src/main/res']
    }
}
```

### 4. proguard-rules.pro Updates

Updated ProGuard rules to keep the new classes:
```
-keep class com.venture.game.MainActivity { *; }
-keep class mobile.EbitenView { *; }
```

Removed obsolete rule:
```
-keep class org.ebitengine.gomobile.GoNativeActivity { *; }
```

### 5. .gitignore Updates

Added Java source directory to allowed paths:
```
!build/android/src/main/java/
!build/android/src/main/java/**
```

## Files Changed

1. **Created:**
   - `build/android/src/main/java/com/venture/game/MainActivity.java` - New custom activity

2. **Modified:**
   - `build/android/AndroidManifest.xml` - Updated activity reference
   - `build/android/build.gradle` - Added Java source directory
   - `build/android/proguard-rules.pro` - Updated ProGuard rules
   - `.gitignore` - Allowed Java sources
   - `docs/ANDROID_BUILD.md` - Updated documentation
   - `ANDROID_BUILD_REPORT.md` - Updated build report

## How It Works

1. **Build Process:**
   - `ebitenmobile bind` creates an AAR with `EbitenView` and `mobile.Mobile` package
   - Gradle compiles `MainActivity.java` and links it with the AAR
   - APK is assembled with both Java code and native Go libraries

2. **Runtime Flow:**
   - Android launches `MainActivity`
   - `onCreate()` initializes Go runtime with `Seq.setContext()`
   - `EbitenView` is created and set as the content view
   - `mobile.Mobile.start()` initializes the game (calls `initializeGame()` in `cmd/mobile/mobile.go`)
   - Game renders through `EbitenView`
   - Lifecycle events are handled by MainActivity

3. **Architecture:**
   ```
   Android Framework
        ↓
   MainActivity (Java)
        ↓
   EbitenView (from AAR)
        ↓
   Go Mobile Runtime (Seq)
        ↓
   Ebiten Game (cmd/mobile/mobile.go)
        ↓
   Game Systems (pkg/engine/*)
   ```

## Why This Approach

This is the **recommended pattern for Ebiten v2.9+** because:

1. **Flexibility**: Full control over the Android Activity lifecycle
2. **Integration**: Easy to add Android-specific features (ads, in-app purchases, etc.)
3. **Standard Practice**: Follows Android development best practices
4. **Future-Proof**: Aligns with ebitenmobile's current and future architecture

The older `GoNativeActivity` approach was a simplified pattern used in earlier versions but has been replaced by the EbitenView pattern in modern Ebiten releases.

## Testing

To verify the fix:

1. Build the AAR:
   ```bash
   ebitenmobile bind -target android -javapkg com.venture.game \
     -o build/android/libs/mobile.aar -androidapi 21 ./cmd/mobile
   ```

2. Build the APK:
   ```bash
   cd build/android && bash gradlew assembleDebug
   ```

3. Install on device:
   ```bash
   adb install -r build/android/build/outputs/apk/debug/android-debug.apk
   ```

4. Launch the app - it should start without the ClassNotFoundException

## References

- Ebiten Mobile Documentation: https://ebitengine.org/en/documents/mobile.html
- gomobile Documentation: https://pkg.go.dev/golang.org/x/mobile
- Android Activity Lifecycle: https://developer.android.com/guide/components/activities/activity-lifecycle
