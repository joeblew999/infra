package dep

import "testing"

func TestCheckAllReleases(t *testing.T) {
	err := CheckAllReleases()
	if err != nil {
		t.Errorf("CheckAllReleases failed: %v", err)
	}
}

func TestCheckGitHubRelease(t *testing.T) {
	release, err := CheckGitHubRelease("oven-sh", "bun")
	if err != nil {
		t.Errorf("CheckGitHubRelease failed: %v", err)
	}
	
	if release.TagName == "" {
		t.Error("Expected non-empty tag name")
	}
	
	t.Logf("Latest bun release: %s", release.TagName)
}