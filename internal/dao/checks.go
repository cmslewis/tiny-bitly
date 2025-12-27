package dao

import (
	"tiny-bitly/internal/dao/memory"
)

// Compile-time interface satisfaction checks.
// These ensure implementations match their interfaces without creating import cycles.
var (
	_ URLRecordDAO = (*memory.URLRecordMemoryDAO)(nil)
)
