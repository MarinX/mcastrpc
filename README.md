# Multicast JSON RPC Server

# Description
Multicast JSON RPC server

# Installation
    go get github.com/MarinX/mcastrpc
# Notes
* Multicast cannot be used in any sort of cloud, or shared infrastructure environment
* Tested only on Linux
* Be sure to run client on other PC, because address on network cannot be same


# Example
    package main

    import (
	    "fmt"
	    "github.com/MarinX/mcastrpc"
    )

    type Api struct {
    }

    type Result struct {
	    Success bool
	    Message string
    }

    func main() {

	    srv := mcastrpc.NewServer()

	    err := srv.Register(new(Api), "Api")
	    if err != nil {
		    fmt.Println(err)
		    return
	    }

        if err := srv.ListenAndServe("224.0.0.251", 1712); err != nil {
		    fmt.Println(err)
	    }

    }

    func (t *Api) Say(r *string, w *Result) error {
	    *w = Result{
		    Success: true,
		    Message: "Hello," + *r,
	    }
	    return nil
    }

# TODO
* Create multicast client


# License
This library is under the MIT License

# Author
Marin Basic 
