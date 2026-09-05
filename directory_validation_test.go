package dirly

import (
	"strings"
	"testing"
)

func TestIsValidName_EmptyString(t *testing.T) {
	err := isValidName("", false)
	if err != nil {
		t.Errorf("expected empty string to be valid, got error: %v", err)
	}
}

func TestIsValidName_SingleCharAlphanumeric(t *testing.T) {
	validChars := []string{"a", "z", "A", "Z", "0", "9"}
	for _, ch := range validChars {
		t.Run(ch, func(t *testing.T) {
			err := isValidName(ch, false)
			if err != nil {
				t.Errorf("expected %q to be valid, got error: %v", ch, err)
			}
		})
	}
}

func TestIsValidName_SingleCharSpecial(t *testing.T) {
	validChars := []string{
		"_", "-", "+", "~", "!", "@",
		"[", "]", "(", ")", "$", "%", "#", "|",
	}
	for _, ch := range validChars {
		t.Run(ch, func(t *testing.T) {
			err := isValidName(ch, false)
			if err != nil {
				t.Errorf("expected %q to be valid, got error: %v", ch, err)
			}
		})
	}
}

func TestIsValidName_SingleCharEuro(t *testing.T) {
	err := isValidName("€", false)
	if err != nil {
		t.Errorf("expected € to be valid, got error: %v", err)
	}
}

func TestIsValidName_MultipleValidCombined(t *testing.T) {
	validNames := []string{
		"abc123",
		"ABC_DEF",
		"my-file-123",
		"test+value",
		"var~name",
		"prefix!suffix",
		"at@sign",
		"brackets[here]",
		"parens(are)ok",
		"dollar$money",
		"percent%done",
		"hash#tag",
		"pipe|value",
		"all_special_-+_~!@#$%|",
		"a1B2c3D4e5F6",
		"mixed-case_123-name",
	}
	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			err := isValidName(name, false)
			if err != nil {
				t.Errorf("expected %q to be valid, got error: %v", name, err)
			}
		})
	}
}

func TestIsValidName_AllSpecialCharsTogether(t *testing.T) {
	valid := "_-+~!@[]()$%#|"
	err := isValidName(valid, false)
	if err != nil {
		t.Errorf("expected all special chars to be valid, got error: %v", err)
	}
}

func TestIsValidName_DotSingle_Valid(t *testing.T) {
	validDotNames := []string{
		"a.b",
		"file.txt",
		"my.config",
	}
	for _, name := range validDotNames {
		t.Run(name, func(t *testing.T) {
			err := isValidName(name, false)
			if err != nil {
				t.Errorf("expected %q to be valid, got error: %v", name, err)
			}
		})
	}
}

func TestIsValidName_DotRules_StartsWithDot(t *testing.T) {
	invalidDotNames := []string{
		".hidden",
		".config",
		".gitignore",
		".",
	}
	for _, name := range invalidDotNames {
		t.Run(name, func(t *testing.T) {
			err := isValidName(name, false)
			if err == nil {
				t.Errorf("expected %q to be invalid (starts with dot), but got no error", name)
			}
		})
	}
}

func TestIsValidName_DotRules_EndsWithDot(t *testing.T) {
	invalidDotNames := []string{
		"file.",
		"test.",
		"a.",
	}
	for _, name := range invalidDotNames {
		t.Run(name, func(t *testing.T) {
			err := isValidName(name, false)
			if err == nil {
				t.Errorf("expected %q to be invalid (ends with dot), but got no error", name)
			}
		})
	}
}

func TestIsValidName_DotRules_MultipleDots(t *testing.T) {
	invalidDotNames := []string{
		"file.txt.bak",
		"my.config.yaml",
		"a.b.c.d",
		"test..double",
	}
	for _, name := range invalidDotNames {
		t.Run(name, func(t *testing.T) {
			err := isValidName(name, false)
			if err == nil {
				t.Errorf("expected %q to be invalid (multiple dots), but got no error", name)
			}
		})
	}
}

func TestIsValidName_GlobDotBehavior(t *testing.T) {
	validGlobNames := []string{
		"file.txt.bak",
		"..double..dots",
		"a.b.c.d.e",
		".",
		"..",
		"...triple",
		"file.",
		".hidden",
	}
	for _, name := range validGlobNames {
		t.Run(name, func(t *testing.T) {
			err := isValidName(name, true)
			if err != nil {
				t.Errorf("expected glob %q to be valid, got error: %v", name, err)
			}
		})
	}
}

func TestIsValidName_GlobStillRejectsInvalidChars(t *testing.T) {
	invalidGlobNames := []string{
		"hello world",  // space
		"hello\\world", // backslash
		"hello:world",  // colon
		"héllo",        // accented
		"日本語",          // Japanese
		"Привет",       // Cyrillic
		"🎉party",       // emoji
	}
	for _, name := range invalidGlobNames {
		t.Run(name, func(t *testing.T) {
			err := isValidName(name, true)
			if err == nil {
				t.Errorf("expected glob %q to be invalid, but got no error", name)
			}
		})
	}
}

func TestIsValidName_EuroAnywhere(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"price-€", false},
		{"€100", false},
		{"test€value", false},
		{"€", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := isValidName(tt.name, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("isValidName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestIsValidName_InvalidChars(t *testing.T) {
	invalidNames := []string{
		"hello world",  // space
		"hello\tworld", // tab
		"hello\nworld", // newline
		"hello/world",  // forward slash
		"hello\\world", // backslash
		"hello:world",  // colon
		"hello*world",  // asterisk
		"hello?world",  // question mark
		"hello\"world", // double quote
		"hello'world",  // single quote
		"hello<world>", // angle brackets
		"hello&world",  // ampersand
		"hello^world",  // caret
		"hello=world",  // equals
	}
	for _, name := range invalidNames {
		t.Run(name, func(t *testing.T) {
			err := isValidName(name, false)
			if err == nil {
				t.Errorf("expected %q to be invalid, but got no error", name)
			}
		})
	}
}

func TestIsValidName_UnicodeInvalid(t *testing.T) {
	invalidUnicode := []string{
		"héllo",  // accented e
		"日本語",    // Japanese
		"Привет", // Cyrillic
		"مرحبا",  // Arabic
		"שלום",   // Hebrew
		"🎉party", // emoji
		"café",   // accented a
		"naïve",  // diaeresis i
	}
	for _, name := range invalidUnicode {
		t.Run(name, func(t *testing.T) {
			err := isValidName(name, false)
			if err == nil {
				t.Errorf("expected %q to be invalid (unicode), but got no error", name)
			}
		})
	}
}

func TestIsValidName_SpecialCharsBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"_start", false},
		{"end_", false},
		{"_middle_", false},
		{"-start", false},
		{"end-", false},
		{"+start", false},
		{"end+", false},
		{"~start", false},
		{"end~", false},
		{"!start", false},
		{"end!", false},
		{"@start", false},
		{"end@", false},
		{"$start", false},
		{"end$", false},
		{"%start", false},
		{"end%", false},
		{"#start", false},
		{"end#", false},
		{"|start", false},
		{"end|", false},
		{"[start", false},
		{"end]", false},
		{"(start", false},
		{"end)", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := isValidName(tt.name, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("isValidName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestIsValidName_MixedValidInvalid(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"valid-name_123", false},
		{"has space", true},
		{"has/slash", true},
		{"has:colon", true},
		{"has*star", true},
		{"has?quest", true},
		{"has\"quote", true},
		{"has'apostrophe", true},
		{"has<angle>", true},
		{"has&amp", true},
		{"has^caret", true},
		{"has=equal", true},
		{"has(paren)ok", false},
		{"has[bracket]ok", false},
		{"has#hash", false},
		{"has!bang", false},
		{"has|pipe", false},
		{"has@at", false},
		{"has$dollar", false},
		{"has%percent", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := isValidName(tt.name, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("isValidName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestIsValidName_VeryLongName(t *testing.T) {
	longName := strings.Repeat("a", 1000)
	err := isValidName(longName, false)
	if err != nil {
		t.Errorf("expected 1000-char valid name to pass, got error: %v", err)
	}
}

func TestIsValidName_ErrorMessage(t *testing.T) {
	err := isValidName("hello world", false)
	if err == nil {
		t.Fatal("expected error for invalid character")
	}
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

func TestIsValidName_GlobVsNonGlobDot(t *testing.T) {
	dotCases := []string{".", "..", "...", "file.", ".hidden", "a.b.c"}
	for _, name := range dotCases {
		t.Run(name, func(t *testing.T) {
			errNonGlob := isValidName(name, false)
			errGlob := isValidName(name, true)
			if errNonGlob == nil {
				t.Errorf("expected %q to be invalid in non-glob mode", name)
			}
			if errGlob != nil {
				t.Errorf("expected %q to be valid in glob mode, got error: %v", name, errGlob)
			}
		})
	}
}

func TestIsValidName_Alphanumeric(t *testing.T) {
	tests := []string{
		"a", "z", "A", "Z", "0", "9",
		"abc123", "ABC123", "a1b2c3",
		"test123TEST", "HelloWorld123",
	}
	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			err := isValidName(name, false)
			if err != nil {
				t.Errorf("expected %q to be valid, got error: %v", name, err)
			}
		})
	}
}

func TestIsValidName_GlobEmptyString(t *testing.T) {
	err := isValidName("", true)
	if err != nil {
		t.Errorf("expected empty string with glob=true to be valid, got error: %v", err)
	}
}

func TestIsValidName_GlobSingleDot(t *testing.T) {
	err := isValidName(".", true)
	if err != nil {
		t.Errorf("expected single dot with glob=true to be valid, got error: %v", err)
	}
}

func TestIsValidName_SpecialCharsOnly(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"---", false},
		{"___", false},
		{"+++", false},
		{"~~~", false},
		{"!!!", false},
		{"@@@", false},
		{"$$$", false},
		{"%%%", false},
		{"###", false},
		{"|||", false},
		{"[]()", false},
		{"-+_~!@#$%|", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := isValidName(tt.name, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("isValidName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestIsValidName_DotWithSpecialChars(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"._", true},   // starts with dot
		{".-", true},   // starts with dot
		{"_.", true},   // ends with dot
		{"-.", true},   // ends with dot
		{"a._", false}, // dot in middle (not at start/end) - valid
		{"a.-", false}, // dot in middle - valid
		{"_.a", false}, // dot in middle - valid
		{"-.a", false}, // dot in middle - valid
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := isValidName(tt.name, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("isValidName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestIsValidName_DoubleDotNonGlob(t *testing.T) {
	err := isValidName("..", false)
	if err == nil {
		t.Error("expected .. with non-glob to be invalid")
	}
}
