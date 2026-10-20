package selfcheck

// FaultMarkers returns the strings only a build with the tag bonsai_test_fault holds: the guard's test fault switch's
// variable, its words and rules, and a function of it (internal/guard/fault_on.go). Their one home, read by two tests:
// cmd/bonsai's TestNormalBuildHasNoFaultCode (none in a normal build, each in a fault build, so the check can see them)
// and this package's TestDist (none in any program a release archive holds). Each is split by + so that no source
// file holds the marker whole, and a search of the tree finds only the code that is the marker.
func FaultMarkers() []string {
	return []string{"BONSAI_TEST" + "_FAULT", "fault-" + "crash", "fault-" + "unknown", "minimal" + "-path",
		"test fault " + "crash", "internal/guard." + "crashHard"}
}
