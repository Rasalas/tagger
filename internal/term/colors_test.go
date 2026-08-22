package term

import "testing"

func TestColorEnabledForceColorWins(t *testing.T) {
	t.Setenv("FORCE_COLOR", "1")
	t.Setenv("NO_COLOR", "1")
	if !colorEnabled() {
		t.Error("FORCE_COLOR should enable colors even with NO_COLOR set")
	}
}

func TestColorDisabledByNoColor(t *testing.T) {
	t.Setenv("FORCE_COLOR", "")
	t.Setenv("NO_COLOR", "1")
	if colorEnabled() {
		t.Error("NO_COLOR should disable colors")
	}
}

func TestDisableClearsAllSequences(t *testing.T) {
	Bold = "\033[1m"
	Primary = "\033[38;2;45;212;191m"
	Yellow = "\033[38;2;251;191;36m"

	disable()

	for name, val := range map[string]string{
		"Bold":    Bold,
		"Dim":     Dim,
		"Reset":   Reset,
		"Primary": Primary,
		"Muted":   Muted,
		"Green":   Green,
		"Red":     Red,
		"Yellow":  Yellow,
	} {
		if val != "" {
			t.Errorf("%s = %q, want empty", name, val)
		}
	}
}
