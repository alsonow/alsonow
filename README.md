# alsonow

[![Go Reference](https://pkg.go.dev/badge/github.com/alsonow/alsonow.svg)](https://pkg.go.dev/github.com/alsonow/alsonow)

**A go web framework.** 🌠

## Install

```bash
go get github.com/alsonow/alsonow
```

## Quick Start

```go
package main

import "github.com/yourusername/alsonow"

func main() {
    an := alsonow.New()

    app.GET("/", func(c *alsonow.Context) {
        c.String(200, "Hello from AlsoNow! 🌠")
    })

    app.Run()
}
```


## License

This project is under the MIT License. See the [LICENSE](LICENSE) file for the full license text.