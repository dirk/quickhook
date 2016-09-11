package main

import (
    "github.com/dirk/quickhook/context"
)

func Install(c *context.Context, prompt bool) error {
    hooks, err := c.ListHooks()
    if err != nil { return err }

    for _, hook := range hooks {
        c.InstallShim(hook, prompt)
    }

    return nil
}
