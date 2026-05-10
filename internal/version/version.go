package version

import (
	_ "embed"
	"strings"
)

//go:embed version.txt
var raw string

// 对外暴露“全局只读变量”
var Version = strings.TrimSpace(raw)
