package saveload

import "testing"

// TestValidateSaveName tests the shared ValidateSaveName function.
func TestValidateSaveName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid simple name",
			input:   "mysave",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			input:   "save123",
			wantErr: false,
		},
		{
			name:    "valid with underscores",
			input:   "my_save_game",
			wantErr: false,
		},
		{
			name:    "valid with dashes",
			input:   "my-save-game",
			wantErr: false,
		},
		{
			name:    "valid with extension",
			input:   "mysave.sav",
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: true,
		},
		{
			name:    "path separator forward slash",
			input:   "../etc/passwd",
			wantErr: true,
		},
		{
			name:    "path separator backslash",
			input:   "..\\windows\\system32",
			wantErr: true,
		},
		{
			name:    "invalid char colon",
			input:   "save:file",
			wantErr: true,
		},
		{
			name:    "invalid char pipe",
			input:   "save|file",
			wantErr: true,
		},
		{
			name:    "invalid char asterisk",
			input:   "save*file",
			wantErr: true,
		},
		{
			name:    "invalid char question mark",
			input:   "save?file",
			wantErr: true,
		},
		{
			name:    "invalid char less than",
			input:   "<save>",
			wantErr: true,
		},
		{
			name:    "invalid char double quote",
			input:   "save\"file",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSaveName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSaveName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
