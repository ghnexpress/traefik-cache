package log

import (
	"fmt"
	"os"
)

func Log(requestID, text string) {
	os.Stdout.WriteString(fmt.Sprintf("[cache-middleware-plugin] [%s] %s\n", requestID, text))
}
