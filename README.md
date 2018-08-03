[![Build Status](https://travis-ci.org/onego-project/xmlrpc.svg?branch=master)](https://travis-ci.org/onego-project/xmlrpc)

# xmlrpc
XMLRPC Go library is designed to make it easy to create an XMLRPC client, making method calls and receiving method responses.

The library is compliant with the XML-RPC specification
published by http://www.xmlrpc.org/. 

## Requirements 
* Go 1.9 or newer to compile
* [Go dep](https://github.com/golang/dep) tool to manage dependencies
* [gometalinter](https://github.com/alecthomas/gometalinter) tool to run Go lint tools and normalise their output 

## Installation 
The recommended way to install this library is using `go get`:
```
go get -u github.com/onego-project/xmlrpc
```

## Usage examples 
Short usage example expects xml-rpc server running at `localhost:8000` with method `pow` which takes 2 integers and returns one int value.
```
package main

import (
	"github.com/onego-project/xmlrpc"
	"context"
	"fmt"
)

func main() {
	client := xmlrpc.NewClient("http://localhost:8000/")

	ctx := context.TODO()

	result, err := client.Call(ctx, "pow", 3, 4)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	fmt.Println("Result kind:", result.Kind())
	fmt.Println("Result:", result.ResultInt())
}
```

## Contributing
1. [Fork xmlrpc library](https://github.com/onego-project/xmlrpc/fork)
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request


