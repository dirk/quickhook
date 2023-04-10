package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/samber/lo"

	"github.com/dirk/quickhook/repo"
)

func install(repo *repo.Repo, quickhook string, prompt bool) error {
	hooks, err := listHooks(repo)
	if err != nil {
		return err
	}

	for _, hook := range hooks {
		shimPath := path.Join(".git", "hooks", hook)
		if prompt {
			shouldInstall, err := promptForInstallShim(os.Stdin, repo, shimPath)
			if err != nil {
				return err
			}

			if !shouldInstall {
				fmt.Printf("Skipping installing shim %v\n", shimPath)
				continue
			}
		}

		installShim(repo, shimPath, quickhook, hook, prompt)

		fmt.Printf("Installed shim %v\n", shimPath)
	}

	return nil
}

func listHooks(repo *repo.Repo) ([]string, error) {
	hooksPath := path.Join(repo.Root, ".quickhook")

	entries, err := ioutil.ReadDir(hooksPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Missing hooks directory: %v\n", hooksPath)
			os.Exit(66) // EX_NOINPUT
		} else {
			return nil, err
		}
	}

	var hooks []string
	for _, entry := range entries {
		name := entry.Name()
		// Rename the mutating hook to the regular pre-commit one.
		if name == "pre-commit-mutating" {
			name = "pre-commit"
		}
		isHook := name == "pre-commit" ||
			name == "commit-msg"
		if entry.IsDir() && isHook {
			hooks = append(hooks, name)
		}
	}
	return lo.Uniq(hooks), nil
}

func promptForInstallShim(stdin io.Reader, repo *repo.Repo, shimPath string) (bool, error) {
	exists, err := repo.IsDir(shimPath)
	if err != nil {
		return false, err
	}

	var message string
	if exists {
		message = fmt.Sprintf("Overwrite existing file %v?", shimPath)
	} else {
		message = fmt.Sprintf("Create file %v?", shimPath)
	}

	scanner := bufio.NewScanner(stdin)
	for {
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
}

func installShim(repo *repo.Repo, shimPath, quickhook, hook string, prompt bool) error {
	command, err := shimCommandForHook(quickhook, hook)
	if err != nil {
		return err
	}

	lines := []string{
		"#!/bin/sh",
		command,
		"", // So we get a trailing newline when we join
	}

	file, err := os.Create(shimPath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = os.Chmod(shimPath, 0755)
	if err != nil {
		return err
	}

	file.WriteString(strings.Join(lines, "\n"))

	return nil
}

func shimCommandForHook(quickhook, hook string) (string, error) {
	var args string

	switch hook {
	case "pre-commit":
		args = "pre-commit"
	case "commit-msg":
		args = "commit-msg $1"
	default:
		return "", fmt.Errorf("invalid hook: %v", hook)
	}

	return fmt.Sprintf("%s hook %s", quickhook, args), nil
}
