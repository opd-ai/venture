#!/bin/bash
set -e

# Generate placeholder Android launcher icons
# This creates simple solid-color placeholder icons for the app

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
RES_DIR="$PROJECT_ROOT/build/android/res"

# Colors
GREEN='\033[0;32m'
NC='\033[0m'

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

# Create directories for different densities
create_mipmap_dirs() {
    echo_info "Creating mipmap directories..."
    mkdir -p "$RES_DIR/mipmap-mdpi"
    mkdir -p "$RES_DIR/mipmap-hdpi"
    mkdir -p "$RES_DIR/mipmap-xhdpi"
    mkdir -p "$RES_DIR/mipmap-xxhdpi"
    mkdir -p "$RES_DIR/mipmap-xxxhdpi"
    mkdir -p "$RES_DIR/mipmap-anydpi-v26"
}

# Generate a simple PNG icon using ImageMagick (if available) or create XML drawable
generate_png_icons() {
    if command -v convert &> /dev/null; then
        echo_info "Generating PNG icons with ImageMagick..."
        
        # mdpi: 48x48
        convert -size 48x48 xc:'#4A90E2' -fill white -gravity center \
            -pointsize 32 -annotate +0+0 'V' \
            "$RES_DIR/mipmap-mdpi/ic_launcher.png"
        
        # hdpi: 72x72
        convert -size 72x72 xc:'#4A90E2' -fill white -gravity center \
            -pointsize 48 -annotate +0+0 'V' \
            "$RES_DIR/mipmap-hdpi/ic_launcher.png"
        
        # xhdpi: 96x96
        convert -size 96x96 xc:'#4A90E2' -fill white -gravity center \
            -pointsize 64 -annotate +0+0 'V' \
            "$RES_DIR/mipmap-xhdpi/ic_launcher.png"
        
        # xxhdpi: 144x144
        convert -size 144x144 xc:'#4A90E2' -fill white -gravity center \
            -pointsize 96 -annotate +0+0 'V' \
            "$RES_DIR/mipmap-xxhdpi/ic_launcher.png"
        
        # xxxhdpi: 192x192
        convert -size 192x192 xc:'#4A90E2' -fill white -gravity center \
            -pointsize 128 -annotate +0+0 'V' \
            "$RES_DIR/mipmap-xxxhdpi/ic_launcher.png"
        
        echo_info "PNG icons generated successfully"
    else
        echo_info "ImageMagick not found, creating XML drawable icons..."
        create_xml_icons
    fi
}

# Create XML drawable icons as fallback
create_xml_icons() {
    for density in mdpi hdpi xhdpi xxhdpi xxxhdpi; do
        cat > "$RES_DIR/mipmap-$density/ic_launcher.xml" <<'EOF'
<?xml version="1.0" encoding="utf-8"?>
<layer-list xmlns:android="http://schemas.android.com/apk/res/android">
    <item>
        <shape android:shape="rectangle">
            <solid android:color="#4A90E2"/>
        </shape>
    </item>
    <item android:gravity="center">
        <shape android:shape="rectangle">
            <solid android:color="#FFFFFF"/>
            <size android:width="48dp" android:height="48dp"/>
        </shape>
    </item>
</layer-list>
EOF
    done
    echo_info "XML drawable icons created"
}

# Create adaptive icon for Android 8.0+
create_adaptive_icon() {
    echo_info "Creating adaptive icon resources..."
    
    mkdir -p "$RES_DIR/drawable"
    
    # Background layer
    cat > "$RES_DIR/drawable/ic_launcher_background.xml" <<'EOF'
<?xml version="1.0" encoding="utf-8"?>
<shape xmlns:android="http://schemas.android.com/apk/res/android"
    android:shape="rectangle">
    <solid android:color="#4A90E2"/>
</shape>
EOF
    
    # Foreground layer
    cat > "$RES_DIR/drawable/ic_launcher_foreground.xml" <<'EOF'
<?xml version="1.0" encoding="utf-8"?>
<vector xmlns:android="http://schemas.android.com/apk/res/android"
    android:width="108dp"
    android:height="108dp"
    android:viewportWidth="108"
    android:viewportHeight="108">
    <group android:scaleX="0.7"
        android:scaleY="0.7"
        android:translateX="16.2"
        android:translateY="16.2">
        <path
            android:fillColor="#FFFFFF"
            android:pathData="M54,0L54,108L0,108L0,0L54,0ZM48,48L6,48L6,102L48,102L48,48Z"/>
    </group>
</vector>
EOF
    
    # Adaptive icon definition
    cat > "$RES_DIR/mipmap-anydpi-v26/ic_launcher.xml" <<'EOF'
<?xml version="1.0" encoding="utf-8"?>
<adaptive-icon xmlns:android="http://schemas.android.com/apk/res/android">
    <background android:drawable="@drawable/ic_launcher_background"/>
    <foreground android:drawable="@drawable/ic_launcher_foreground"/>
</adaptive-icon>
EOF
    
    echo_info "Adaptive icon created"
}

# Main execution
main() {
    echo_info "Generating Android launcher icons..."
    
    create_mipmap_dirs
    generate_png_icons
    create_adaptive_icon
    
    echo_info "Icon generation complete!"
}

main "$@"
