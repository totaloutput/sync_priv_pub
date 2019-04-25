package tools

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

// AskUser asks the user if it's ok to continue, and returns true if so
func AskUser() (bool, error) {
	for {
		fmt.Print("Ok to continue? (y\\n): ")
		in := bufio.NewReader(os.Stdin)
		answer, err := in.ReadString('\n')
		if err != nil {
			return false, err
		}

		matched, err := regexp.MatchString("(?i)y(?:es)?", answer)
		if err != nil {
			return false, err
		}

		if matched {
			return true, nil
		}

		matched, err = regexp.MatchString("(?i)no?", answer)
		if err != nil {
			return false, err
		}

		if matched {
			return false, nil
		}
	}
}

