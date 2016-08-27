package hooks

import (
	"fmt"
)

func PreCommit() error {
	fmt.Println("Hello world!")
	return nil
}
