# Android Build Configuration

This directory contains the Android build configuration for Venture.

## Structure

```
build/android/
├── AndroidManifest.xml    # App manifest with permissions and activities
├── build.gradle          # Gradle build configuration
├── settings.gradle       # Gradle settings
├── README.md            # This file
└── res/                 # Android resources
    ├── drawable/        # Vector drawables
    │   ├── ic_launcher_background.xml
    │   └── ic_launcher_foreground.xml
    ├── mipmap-*/        # Launcher icons (various densities)
    │   └── ic_launcher.png
    ├── mipmap-anydpi-v26/  # Adaptive icon (Android 8.0+)
    │   └── ic_launcher.xml
    └── values/
        └── strings.xml  # String resources
```

## Building

Use the `scripts/build-android.sh` script to build the Android app:

```bash
# Build AAR library only
./scripts/build-android.sh aar

# Build debug APK
./scripts/build-android.sh apk

# Build release APK (requires signing configuration)
./scripts/build-android.sh apk-release

# Build Android App Bundle for Play Store
./scripts/build-android.sh aab

# Build and install on connected device
./scripts/build-android.sh install
```

## Resources

### Launcher Icons

Launcher icons are generated automatically by the `scripts/generate-android-icons.sh` script, which is called during the build process. The icons are created in multiple densities:

- **mdpi**: 48x48px
- **hdpi**: 72x72px
- **xhdpi**: 96x96px
- **xxhdpi**: 144x144px
- **xxxhdpi**: 192x192px

The script uses ImageMagick to generate simple placeholder icons with a blue background and white "V" letter. You can customize these by:

1. Creating your own icons in the appropriate sizes
2. Placing them in the `res/mipmap-*` directories
3. Overwriting the generated files

### Adaptive Icons

For Android 8.0 (API 26) and above, the app uses adaptive icons defined in:
- `res/mipmap-anydpi-v26/ic_launcher.xml` - Main adaptive icon definition
- `res/drawable/ic_launcher_background.xml` - Background layer
- `res/drawable/ic_launcher_foreground.xml` - Foreground layer

Adaptive icons allow the system to display the icon in different shapes (circle, square, rounded square) depending on the device manufacturer's theme.

## Requirements

- Go 1.24+
- Android SDK (API 34)
- Android NDK (26.1.10909125)
- ebitenmobile (`go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest`)
- ImageMagick (optional, for icon generation)

## Configuration

### Package Name

The app package name is defined in `AndroidManifest.xml`:
```xml
package="com.venture.game"
```

### Minimum SDK Version

Minimum Android version: API 21 (Android 5.0 Lollipop)

### Target SDK Version

Target Android version: API 34 (Android 14)

### Permissions

The app requires the following permissions:
- `INTERNET` - For multiplayer networking
- `ACCESS_NETWORK_STATE` - To check network availability
- `VIBRATE` - For haptic feedback
- `WAKE_LOCK` - To keep screen on during gameplay

### Graphics

The app requires OpenGL ES 2.0 for rendering.

## Troubleshooting

### Icons Not Found Error

If you see an error like:
```
AAPT: error: resource mipmap/ic_launcher not found
```

Run the icon generation script manually:
```bash
./scripts/generate-android-icons.sh
```

### ImageMagick Not Available

If ImageMagick is not installed, the icon generation script will fall back to creating XML drawable icons. To install ImageMagick:

```bash
# Ubuntu/Debian
sudo apt-get install imagemagick

# macOS
brew install imagemagick
```

### Gradle Build Failures

Make sure your ANDROID_HOME and ANDROID_NDK_HOME environment variables are set correctly:

```bash
export ANDROID_HOME=$HOME/Android/Sdk
export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/26.1.10909125
```

### Java Version Issues

The build requires JDK 17. If you have multiple Java versions installed, set JAVA_HOME:

```bash
export JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64
```

## CI/CD

The GitHub Actions workflow (`.github/workflows/android.yml`) automatically builds the Android APK on pushes to main/develop branches. The workflow:

1. Sets up Go, Java 17, and Android SDK
2. Installs NDK and ebitenmobile
3. Generates launcher icons (with ImageMagick)
4. Builds the AAR library
5. Builds debug APK (or release if triggered manually)
6. Uploads artifacts

## Customization

### Changing App Name

Edit `res/values/strings.xml`:
```xml
<string name="app_name">Venture</string>
```

### Changing Package Name

1. Update `AndroidManifest.xml` package attribute
2. Update `PACKAGE_NAME` in `scripts/build-android.sh`
3. Update `build.gradle` namespace

### Changing Version

Update in `AndroidManifest.xml`:
```xml
android:versionCode="1"
android:versionName="1.0.0"
```

And in `scripts/build-android.sh`:
```bash
VERSION_NAME="1.0.0"
VERSION_CODE="1"
```

## Release Builds

For release builds, you need to configure signing:

1. Create a keystore:
   ```bash
   keytool -genkey -v -keystore venture.keystore -alias venture -keyalg RSA -keysize 2048 -validity 10000
   ```

2. Set environment variables:
   ```bash
   export VENTURE_KEYSTORE_FILE=path/to/venture.keystore
   export VENTURE_KEYSTORE_PASSWORD=your_password
   export VENTURE_KEY_ALIAS=venture
   export VENTURE_KEY_PASSWORD=your_key_password
   ```

3. Build release APK:
   ```bash
   ./scripts/build-android.sh apk-release
   ```

For GitHub Actions, store these as repository secrets:
- `ANDROID_KEYSTORE_FILE` (base64 encoded keystore)
- `ANDROID_KEYSTORE_PASSWORD`
- `ANDROID_KEY_ALIAS`
- `ANDROID_KEY_PASSWORD`
