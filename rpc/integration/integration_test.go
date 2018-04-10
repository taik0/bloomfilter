package integration

import (
	"context"
	"fmt"
	"net/rpc"

	"github.com/letgoapp/go-bloomfilter"
	"github.com/letgoapp/go-bloomfilter/rotate"
	bf_rpc "github.com/letgoapp/go-bloomfilter/rpc"
)

func ExampleIntegration() {
	fmt.Println("connecting")
	client, err := rpc.DialHTTP("tcp", "127.0.0.1:1234")
	fmt.Println("connected")
	if err != nil {
		fmt.Printf("dialing error: %s", err.Error())
		return
	}

	var (
		addOutput   bf_rpc.AddOutput
		checkOutput bf_rpc.CheckOutput
		unionOutput bf_rpc.UnionOutput
		elems1      = [][]byte{[]byte("rrrr"), []byte("elem2")}
		elems2      = [][]byte{[]byte("house")}
		elems3      = [][]byte{[]byte("house"), []byte("mouse")}
		cfg         = rotate.Config{
			Config: bloomfilter.Config{
				N:        10000000,
				P:        0.0000001,
				HashName: "optimal",
			},
			TTL: 1500,
		}
	)

	err = client.Call("Bloomfilter.Add", bf_rpc.AddInput{elems1}, &addOutput)
	if err != nil {
		fmt.Printf("unexpected error: %s", err.Error())
		return
	}

	err = client.Call("Bloomfilter.Check", bf_rpc.CheckInput{elems1}, &checkOutput)
	if err != nil {
		fmt.Printf("unexpected error: %s", err.Error())
		return
	}
	if len(checkOutput.Checks) != 2 || !checkOutput.Checks[0] || !checkOutput.Checks[1] {
		fmt.Printf("checks error, expected true elements")
		return
	}

	divCall := client.Go("Bloomfilter.Check", elems1, &checkOutput, nil)
	<-divCall.Done

	if len(checkOutput.Checks) != 2 || !checkOutput.Checks[0] || !checkOutput.Checks[1] {
		fmt.Printf("checks error, expected true elements")
		return
	}

	var bf2 = rotate.New(context.Background(), cfg)
	bf2.Add([]byte("house"))

	err = client.Call("Bloomfilter.Union", bf_rpc.UnionInput{bf2}, &unionOutput)
	if err != nil {
		fmt.Printf("unexpected error: %s", err.Error())
		return
	}
	fmt.Println(unionOutput.Capacity < 1e-6)

	err = client.Call("Bloomfilter.Check", bf_rpc.CheckInput{elems2}, &checkOutput)
	if err != nil {
		fmt.Printf("unexpected error: %s", err.Error())
		return
	}
	fmt.Println(checkOutput.Checks)
	if len(checkOutput.Checks) != 1 || !checkOutput.Checks[0] {
		fmt.Println("checks error, expected true element")
		return
	}

	var bf3 = rotate.New(context.Background(), cfg)
	bf3.Add([]byte("mouse"))

	divCall = client.Go("Bloomfilter.Union", bf_rpc.UnionInput{bf3}, &unionOutput, nil)
	<-divCall.Done

	err = client.Call("Bloomfilter.Check", bf_rpc.CheckInput{elems3}, &checkOutput)
	if err != nil {
		fmt.Printf("unexpected error: %s", err.Error())
		return
	}

	fmt.Println(unionOutput.Capacity < 1e-6)
	fmt.Println(checkOutput.Checks)
	if len(checkOutput.Checks) != 2 || !checkOutput.Checks[0] || !checkOutput.Checks[1] {
		fmt.Println("checks error, expected true element")
		return
	}

	// Output:
	// true
	// [true]
	// true
	// [true true]
}
