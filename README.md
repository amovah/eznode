# eznode

Load balancing and failed requests recovery with respecting to node request rate limit for blockchain nodes.

## Why

When working with a blockchain application, you need to read the blockchain data and process it. blockchain nodes
doens't always respond to requests or doesn't respond properly. Sometimes there is two public node but they have request
rate limit.

For application, it's important to know the node that is responding to the request. With eznode you improve your
application SLA, load balancing between nodes and using public node without thinking about rate limit (you may still
face rate limit if you request more than nodes can handle).

## Features

* Load Balance
* Failed Request Recovery
* Node Request Rate Limit
* Disable/Enable Nodes
* Prioritize Nodes
* Node Performance Statistics

## Usage

Install eznode:

```shell
go get github.com/amovah/eznode
```

Create a go file then:

```go
package main

import (
	"fmt"
	"github.com/amovah/eznode"
	"time"
	"net/http"
)

func main() {
	node1 := eznode.NewChainNode(eznode.ChainNodeData{
		Name: "node 1",
		Url:  "https://example.com",
		Limit: eznode.ChainNodeLimit{
			Count: 10,
			Per:   5 * time.Second,
		},
		RequestTimeout: 10 * time.Second,
		Priority:       1,
		Middleware:     nil, // optional
	})

	node2 := eznode.NewChainNode(eznode.ChainNodeData{
		Name: "node 2",
		Url:  "https://example.com",
		Limit: eznode.ChainNodeLimit{
			Count: 10,
			Per:   5 * time.Second,
		},
		RequestTimeout: 10 * time.Second,
		Priority:       2,
		Middleware:     nil, // optional
	})

	chain := eznode.NewChain(eznode.ChainData{
		Id: "Ethereum",
		Nodes: []*eznode.ChainNode{
			node1,
			node2,
		},
		CheckTickRate: eznode.CheckTick{
			TickRate:         100 * time.Millisecond,
			MaxCheckDuration: 5 * time.Second,
		},
	})

	createdEzNode := eznode.NewEzNode([]*eznode.Chain{chain})

	// sample http request
	req, _ := http.NewRequest("GET", "/latest-block", nil)
	// target ethereum chain
	// eznode will automatically select the node that has the highest priority
	// then will check the node request rate limit
	// if the node is not responding, eznode will try to recover the request
	// and try to send the request to the another node
	response, _ := createdEzNode.SendRequest("Ethereum", req)
	fmt.Println(response)
}
```

## LICENSE

Apache License Version 2.0
