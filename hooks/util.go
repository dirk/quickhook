package hooks

import (
    "github.com/fatih/color"
)

func errToStringStatus(err error) string {
    if err == nil {
        return color.GreenString("ok")
    } else {
        return color.RedString("fail")
    }
}
