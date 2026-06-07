package clipboard

import "fmt"

// PrintPasteSetup writes paste capability status and setup steps to stdout.
func PrintPasteSetup() {
	fmt.Println()
	fmt.Println("Auto-paste check:")
	if PasteReady() {
		fmt.Println("  ✓ ready (backend:", PasteToolName()+")")
		return
	}
	if SessionNeedsRelogin() {
		fmt.Println("  ! input group added but not active in this session")
		fmt.Println()
		fmt.Println(PasteSessionHint())
		fmt.Println()
		return
	}
	fmt.Println("  ✗ not ready on this system")
	fmt.Println()
	fmt.Println(PasteSetupHint())
	fmt.Println()
}

// SetupPasteExitCode returns 0 when paste works, 1 when setup is needed.
func SetupPasteExitCode() int {
	if PasteReady() {
		return 0
	}
	return 1
}
