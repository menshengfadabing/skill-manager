package manager

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm asks the user to type yes. Non-TTY requires force=true (--yes).
func Confirm(prompt string, force bool) error {
	if force {
		return nil
	}
	fi, err := os.Stdin.Stat()
	if err != nil || (fi.Mode()&os.ModeCharDevice) == 0 {
		return fmt.Errorf("非交互环境必须加 --yes 才能执行破坏性操作")
	}
	fmt.Fprintf(os.Stderr, "%s\n输入 yes 继续: ", prompt)
	sc := bufio.NewScanner(os.Stdin)
	if !sc.Scan() {
		return fmt.Errorf("已取消")
	}
	ans := strings.TrimSpace(strings.ToLower(sc.Text()))
	if ans != "yes" && ans != "y" {
		return fmt.Errorf("已取消")
	}
	return nil
}
