package templates

// TextIsland represents a localized content fragment that UI templates can
// render while the deterministic pipeline is still under development.
type TextIsland struct {
	Key    string
	Locale string
	Title  string
	Body   string
}

// testIslands contains short placeholder translations we can use until the
// content system is wired to persistent storage.
var testIslands = []TextIsland{
	{
		Key:    "welcome",
		Locale: "en",
		Title:  "Welcome",
		Body:   "Explore the deterministic core runtime snapshot.",
	},
	{
		Key:    "welcome",
		Locale: "de",
		Title:  "Willkommen",
		Body:   "Erkunde den deterministischen Core-Laufzeit-Snapshot.",
	},
}

// TestIslands returns a copy of the placeholder text islands so callers can
// safely modify the returned slice without affecting the backing store.
func TestIslands() []TextIsland {
	out := make([]TextIsland, len(testIslands))
	copy(out, testIslands)
	return out
}
