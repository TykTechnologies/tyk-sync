urljoin
=======

Join URL parts into a single URL, with correctly added slashes.  
It doesn't add a trailing slash, but keeps one if present.

## Install

```go
go get github.com/ongoingio/urljoin
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/ongoingio/urljoin"
)

func main() {
    url := urljoin.Join("http://example.com", "foo")
    fmt.Println(url) // Output: http://example.com/foo
}
```

## Test

```go
go test
```

## License

MIT
