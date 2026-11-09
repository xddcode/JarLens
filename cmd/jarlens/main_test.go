package main

import "testing"

func TestStartupProjectRestoresMostRecentArchive(t *testing.T) {
	path, restore := startupProject(nil, []string{"last.jar", "older.jar"}, true)
	if path != "last.jar" || !restore {
		t.Fatalf("startup project = %q, restore %v; want last.jar with saved state", path, restore)
	}
}

func TestStartupProjectPrefersExplicitArgument(t *testing.T) {
	path, restore := startupProject([]string{"chosen.jar"}, []string{"last.jar"}, true)
	if path != "chosen.jar" || restore {
		t.Fatalf("startup project = %q, restore %v; want explicit project", path, restore)
	}
}

func TestStartupProjectCanDisableRestore(t *testing.T) {
	path, restore := startupProject(nil, []string{"last.jar"}, false)
	if path != "" || restore {
		t.Fatalf("disabled restoration selected %q, restore %v", path, restore)
	}
}

func TestStartupProjectWithoutHistoryStaysHome(t *testing.T) {
	path, restore := startupProject(nil, nil, true)
	if path != "" || restore {
		t.Fatalf("empty startup selected %q, restore %v", path, restore)
	}
}
