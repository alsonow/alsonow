# alsonow

[![Go Reference](https://pkg.go.dev/badge/github.com/alsonow/alsonow.svg)](https://pkg.go.dev/github.com/alsonow/alsonow)
[![Go version](https://img.shields.io/github/go-mod/go-version/alsonow/alsonow)](https://github.com/alsonow/alsonow)
[![GitHub license](https://img.shields.io/github/license/alsonow/alsonow)](https://github.com/alsonow/alsonow/blob/main/LICENSE)
[![GitHub stars](https://img.shields.io/github/stars/alsonow/alsonow?style=social)](https://github.com/alsonow/alsonow/stargazers)

**A go web framework.** 🌠

## Install

```bash
go get github.com/alsonow/alsonow
```

## Quick Start

```go
package main

import "github.com/alsonow/alsonow"

func main() {
    an := alsonow.New()

    an.GET("/", func(c *alsonow.Context) {
		c.Writer.Write([]byte("Hello from AlsoNow! 🌠"))
    })

    an.Run()
}
```


## License

This project is under the MIT License. See the [LICENSE](LICENSE) file for the full license text.