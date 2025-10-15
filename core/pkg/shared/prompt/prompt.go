package prompt

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// YesNo prompts the user for a yes/no answer
func YesNo(reader *bufio.Reader, out io.Writer, prompt string, defaultYes bool) (bool, error) {
	hint := "[y/N]"
	if defaultYes {
		hint = "[Y/n]"
	}

	for {
		if _, err := fmt.Fprintf(out, "%s %s ", prompt, hint); err != nil {
			return false, err
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}

		resp := strings.TrimSpace(strings.ToLower(line))
		if resp == "" {
			return defaultYes, nil
		}

		switch resp {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			if _, err := fmt.Fprintln(out, "Please answer y or n."); err != nil {
				return false, err
			}
		}
	}
}

// Selection prompts the user to select from a list of options
func Selection(reader *bufio.Reader, out io.Writer, prompt string, count int) (int, error) {
	for {
		if _, err := fmt.Fprintf(out, "%s ", prompt); err != nil {
			return -1, err
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			return -1, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			return -1, nil
		}

		idx, err := strconv.Atoi(line)
		if err != nil || idx < 1 || idx > count {
			if _, err := fmt.Fprintln(out, "Invalid selection; enter a number from the list."); err != nil {
				return -1, err
			}
			continue
		}

		return idx - 1, nil
	}
}

// String prompts the user for a string input
func String(reader *bufio.Reader, out io.Writer, prompt string) (string, error) {
	fmt.Fprintf(out, "%s ", prompt)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
