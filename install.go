package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dirk/quickhook/context"
)

func Install(c *context.Context, prompt bool) error {
	hooks, err := c.ListHooks()
	if err != nil {
		return err
	}

	for _, hook := range hooks {
		shimPath := context.PathForShim(hook)

		if prompt {
			shouldInstall, err := promptForInstallShim(hook, shimPath)
			if err != nil {
				return err
			}

			if !shouldInstall {
				fmt.Printf("Skipping installing shim %v\n", shimPath)
				continue
			}
		}

		c.InstallShim(hook, prompt)

		fmt.Printf("Installed shim %v\n", shimPath)
	}

	return nil
}

func promptForInstallShim(hook string, shimPath string) (bool, error) {
	exists, err := exists(shimPath)
	if err != nil {
		return false, err
	}

	var message string
	if exists {
		message = fmt.Sprintf("Overwrite existing file %v?", shimPath)
	} else {
		message = fmt.Sprintf("Create file %v?", shimPath)
	}

	scanner := bufio.NewScanner(os.Stdin)

	for true {
		fmt.Printf("%v [yn] ", message)

		if !scanner.Scan() {
			return false, scanner.Err()
		}

		reply := strings.ToLower(scanner.Text())

		if len(reply) == 0 {
			continue
		}

		switch reply[0] {
		case 'y':
			return true, nil
		case 'n':
			return false, nil
		default:
			continue
		}
	}

	return false, fmt.Errorf("unreachable")
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false, nil
	}

	return err == nil, err
}
