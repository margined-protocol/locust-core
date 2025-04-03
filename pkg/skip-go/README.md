# Skip Go Client

## Example usage

```go
package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/margined-protocol/locust/core/pkg/skip-go"
)

func main() {
	ctx := context.Background()

	skipgo, err := skipgo.NewClient("https://api.skip.build")
	if err != nil {
		panic(err)
	}

	amount := big.NewInt(1000000)
	route, err := skipgo.Route(
    	ctx, 
	    "uusdc", 
	    "axelar-dojo-1", 
	    "uatom", 
	    "cosmoshub-4", 
	    amount,
  )
	if err != nil {
		panic(err)
	}

	fmt.Printf("Route: %v\n", route)
}
```
