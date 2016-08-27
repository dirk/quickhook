package hooks

import (
	"fmt"

	"github.com/dirk/quickhook/context"
)

func PreCommit(c *context.Context) error {
	files, err := c.FilesToBeCommited()
	if err != nil { return err }

	for _, file := range files {
		fmt.Printf("file: %v\n", file)
	}

	return nil
}
