package com.venture.game;

import android.app.Activity;
import android.os.Build;
import android.os.Bundle;
import android.util.Log;
import android.view.View;
import android.view.WindowInsets;
import android.view.WindowInsetsController;
import android.view.WindowManager;

import go.Seq;
import mobile.EbitenView;

public class MainActivity extends Activity {
    private static final String TAG = "VentureGame";
    private EbitenView ebitenView;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        
        try {
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
        } catch (Exception e) {
            Log.e(TAG, "Failed to initialize game", e);
            // Show error and exit gracefully
            finish();
        }
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
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
            // Use modern WindowInsetsController for Android 11+
            WindowInsetsController controller = getWindow().getInsetsController();
            if (controller != null) {
                controller.hide(WindowInsets.Type.systemBars());
                controller.setSystemBarsBehavior(
                    WindowInsetsController.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE);
            }
        } else {
            // Use deprecated API for older Android versions
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
}
