// Package progress describes preparation stages without inventing byte counts
// or percentages for operations whose progress is not reported by suppliers.
package progress

const (
	Download  = "download"
	Extract   = "extract"
	Configure = "configure"
	Toolchain = "toolchain"
	Publish   = "publish"
)

// Event announces an operation before it starts. Component is its display name;
// Index is one-based and Total counts components, not percent completion. Counts
// are zero for stages that do not operate on an individual component.
type Event struct {
	Stage     string
	Component string
	Index     int
	Total     int
}
