package settings

import (
	"testing"
)

// TestFindAll_IsFn_filtersUnknownSettings بدون panic تنظیمات ناشناخته را حذف می‌کند.
func TestFindAll_IsFn_filtersUnknownSettings(t *testing.T) {
	list := make([]SecuritySetting, 0, len(DefaultSettings))
	for name, value := range DefaultSettings {
		list = append(list, SecuritySetting{SettingName: name, SettingValue: value})
	}

	got := filterSettingsForDisplay(list)

	if len(got) != len(SettingsDicFN) {
		t.Fatalf("expected %d settings, got %d", len(SettingsDicFN), len(got))
	}
	for _, s := range got {
		if _, ok := SettingsDicFN[s.SettingName]; ok {
			// renamed to Persian label — skip
			continue
		}
		found := false
		for _, label := range SettingsDicFN {
			if label == s.SettingName {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("unexpected setting name %q", s.SettingName)
		}
	}
}
