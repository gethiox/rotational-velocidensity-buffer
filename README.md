# Rotational Velocidensity Buffer
[![GoDoc](https://godoc.org/github.com/gethiox/rotational-velocidensity-buffer?status.svg)](https://godoc.org/github.com/gethiox/rotational-velocidensity-buffer)
[![Go Report Card](https://goreportcard.com/badge/github.com/gethiox/rotational-velocidensity-buffer)](https://goreportcard.com/report/github.com/gethiox/rotational-velocidensity-buffer)

RVB is a generic version of classical circular buffer.
Thanks to [generics](https://go.dev/doc/tutorial/generics), user can conveniently define inner type
that will be held by the buffer.

### Minimum version requirement

Due to usage of built-in `min()` and `max()` functions, minimum supported version is **[Go 1.21](https://go.dev/doc/go1.21)**.

### Installation

```shell
go get -v -u github.com/gethiox/rotational-velocidensity-buffer
```

### Example usage

##### simple push and read

```go
package main

import (
    "fmt"
    "time"

    "github.com/gethiox/rotational-velocidensity-buffer"
)

type ThingToStore struct {
    ID   int
    Text string
}

const SleepTime = 100 * time.Millisecond

func main() {
    size := 6
    myBuffer := rvb.NewBuffer[ThingToStore](size)

    go func() {
        var i int
        for {
            time.Sleep(SleepTime)
            myBuffer.Push(ThingToStore{ID: i, Text: "Hello World!"})
            i++
        }
    }()

    // minimize time-sensitivity
    time.Sleep(SleepTime / 2)

    // giving time for 3 insertions
    time.Sleep(SleepTime * 3)

    fmt.Println("ReadNew():")
    output := myBuffer.ReadNew(size)
    for _, thing := range output {
        fmt.Printf("%d: %s\n", thing.ID, thing.Text)
    }
    fmt.Println()

    // giving time for 3 additional insertions, buffer will be entirely populated
    time.Sleep(SleepTime * 3)

    fmt.Println("ReadNew():")
    output = myBuffer.ReadNew(size)
    for _, thing := range output {
        fmt.Printf("%d: %s\n", thing.ID, thing.Text)
    }
    fmt.Println()
    
    // giving time for 3 additional insertions, buffer will be partially overwritten
    time.Sleep(SleepTime * 3)

    fmt.Println("ReadNew():")
    output = myBuffer.ReadNew(size)
    for _, thing := range output {
        fmt.Printf("%d: %s\n", thing.ID, thing.Text)
    }
}
```

###### output

```terminaloutput
ReadNew():
2: Hello World!
1: Hello World!
0: Hello World!

ReadNew():
5: Hello World!
4: Hello World!
3: Hello World!
2: Hello World!
1: Hello World!
0: Hello World!

ReadNew():
8: Hello World!
7: Hello World!
6: Hello World!
5: Hello World!
4: Hello World!
3: Hello World!
```

##### utilizing checkpoint

```go
package main

import (
    "fmt"
    "time"

    "github.com/gethiox/rotational-velocidensity-buffer"
)

type ThingToStore struct {
    ID   int
    Text string
}

const SleepTime = 100 * time.Millisecond

func main() {
    size := 6
    myBuffer := rvb.NewBuffer[ThingToStore](size)

    go func() {
        var i int
        for {
            time.Sleep(SleepTime)
            myBuffer.Push(ThingToStore{ID: i, Text: "Hello World!"})
            i++
        }
    }()

    // minimize time-sensitivity
    time.Sleep(SleepTime / 2)

    // giving time for 3 insertions
    time.Sleep(SleepTime * 3)

    // saving current buffer position
    checkpoint := myBuffer.GetCheckpoint()

    // giving time for 3 additional insertions, buffer will be entirely populated
    time.Sleep(SleepTime * 3)

    fmt.Println("current view from checkpoint:")
    output, missing := myBuffer.ReadNewFromCheckpoint(checkpoint, 0, size)
    for _, thing := range output {
        fmt.Printf("%d: %s\n", thing.ID, thing.Text)
    }
    fmt.Printf("missing: %#v\n", missing)
    fmt.Println()

    // giving time for 2 additional insertions, buffer will be partially overwritten
    time.Sleep(SleepTime * 2)

    fmt.Println("current view from checkpoint:")
    output, missing = myBuffer.ReadNewFromCheckpoint(checkpoint, 0, size)
    for _, thing := range output {
        fmt.Printf("%d: %s\n", thing.ID, thing.Text)
    }
    fmt.Printf("missing: %#v\n", missing)
}
```

###### output

```terminaloutput
current view from checkpoint:
2: Hello World!
1: Hello World!
0: Hello World!
missing: rvb.Missing{Reused:0, Max:3}

current view from checkpoint:
2: Hello World!
missing: rvb.Missing{Reused:2, Max:3}
```
