package util

import (
	"os"
	"strconv"
	"strings"
)

func GetEnv(name string, def ...string) string {
	val := os.Getenv(name)
	if val == "" && len(def) > 0 {
		val = def[0]
	}
	return val
}

func GetEnvInt(name string, def ...int) int {
	if val := os.Getenv(name); val != "" {
		iVal, _ := strconv.Atoi(strings.TrimSpace(val))
		return iVal
	}

	if len(def) > 0 {
		return def[0]
	}
	return 0
}
