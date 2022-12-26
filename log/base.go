package log

import (
	"fmt"
	"os"
)

func Log(text string) {
	os.Stdout.WriteString(fmt.Sprintf("[cache-middleware-plugin] %s\n", text))
}
