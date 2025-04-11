package utils

import (
	"strings"

	"github.com/openkcm/api-sdk/internal/ops"
)

func Contains(s string, str string) bool {
	str = ops.TrimSpaces(str)
	return strings.Contains(s, str)
}
