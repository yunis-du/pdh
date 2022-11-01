package tools

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

func IsEmpty(s string) bool {
	return len(s) == 0
}

func IsBlank(s string) bool {
	return len(strings.Trim(s, " ")) == 0
}

func GenRandStr(n int, join string) string {
	result := ""
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 2)
	for i := 0; i < n; i++ {
		rand.Read(b)
		result += hex.EncodeToString(b)
		if i < n-1 {
			result += join
		}
	}
	return result
}

func ByteCountDecimal(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func GetInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Fprintf(os.Stderr, "%s", prompt)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}
