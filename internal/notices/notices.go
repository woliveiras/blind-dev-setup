// Package notices preserves the license notices shipped with the executable.
package notices

import _ "embed"

//go:embed THIRD-PARTY-NOTICES.txt
var Text string
