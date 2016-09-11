package main

import (
    "github.com/dirk/quickhook/context"
)

func Install(c *context.Context) error {
    hooks, err := c.ListHooks()
    if err != nil { return err }

    for _, hook := range hooks {
        c.InstallShim(hook)
    }

    return nil
}
