//go:build !bonsai_test_fault

package guard

// A normal build: no test fault. The fault switch (plan part 5, spec §14) lives in fault_on.go, which only a build
// with the tag bonsai_test_fault compiles; this file is its stand-in, and it reads no variable and holds no fault.
// cmd/bonsai's TestNormalBuildHasNoFaultCode builds the binary both ways and checks the normal one for any trace of
// the switch.

// FaultBuild reports whether this build holds the test fault switch.
const FaultBuild = false

// testFault is where a fault build acts out its fault. Here it does nothing.
func testFault(*run) *Decision { return nil }
