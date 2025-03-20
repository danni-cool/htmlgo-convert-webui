package main

import (
    "github.com/theplant/htmlgo"
)

func main() {
    // This will fail with "unexpected if, expected expression"
    var n = if true {
        htmlgo.Div().Text("True condition")
    } else {
        htmlgo.Div().Text("False condition")
    }

    // This will never execute
    fmt.Println(htmlgo.MustString(n, nil))
} 