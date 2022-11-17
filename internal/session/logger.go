package session

import (
	"fmt"
	"io"
	"log"
	"strings"
)

func writeLog(w io.Writer, leader, sessionID, line string) {
	line = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(line), "\r", `\r`), "\n", `\n`), "\t", `\t`)

	if len(line) > maxLineLength {
		line = line[:maxLineLength] + "..."
	}

	if _, err := fmt.Fprintf(w, "%v[%v]: %v\n", leader, sessionID, line); err != nil {
		log.Printf("gluon: failed to write log: %v", err)
	}
}
