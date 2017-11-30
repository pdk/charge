package main

// Use skylark to define required inputs, and a charge compute function.

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/google/skylark"
	"github.com/google/skylark/resolve"
)

func init() {
	resolve.AllowNestedDef = true // allow def statements within function bodies
	resolve.AllowLambda = true    // allow lambda expressions
	resolve.AllowFloat = true     // allow floating point literals, the 'float' built-in, and x / y
	resolve.AllowSet = true       // allow the 'set' built-in
}

func main() {
	thread := &skylark.Thread{
		Print: func(_ *skylark.Thread, msg string) { fmt.Println(msg) },
	}

	// Define a callback so the client can define which inputs, types, sources
	// are required to execute the compute function.
	defineRequiredInput := skylark.NewBuiltin("requireInput",
		func(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
			log.Printf("defining required input: %v", args)
			return skylark.String("ack"), nil
		})

	// Put the callback in the script's global env.
	globals := skylark.StringDict{
		"requireInput": defineRequiredInput,
	}

	data, err := ioutil.ReadFile("charge.sky")
	if err != nil {
		log.Fatal(err)
	}

	if err := skylark.ExecFile(thread, "charge.sky", data, globals); err != nil {
		if evalErr, ok := err.(*skylark.EvalError); ok {
			log.Fatal(evalErr.Backtrace())
		}
		log.Fatal(err)
	}

	// Make sure they defined the function
	computeCharge, ok := globals["computeCharge"]
	if !ok {
		log.Fatal("definition of computeCharge is required")
	}

	// Make sure it's actually a function
	computeChargeFn, ok := computeCharge.(skylark.Callable)
	if !ok {
		log.Fatal("computeCharge must be defined as a function")
	}

	// Setup the arguments according to what was registered as inputs.
	var args skylark.Tuple
	args = append(args, skylark.MakeInt(23))
	args = append(args, skylark.MakeInt(1203))
	args = append(args, skylark.MakeInt(300))
	args = append(args, skylark.String("blue"))

	// Execute the client defined function.
	result, err := computeChargeFn.Call(thread, args, nil)
	if err != nil {
		log.Fatal(err)
	}

	if result.Type() != "float" {
		log.Fatalf("expected the compute function to return a float, but got %s", result.Type())
	}

	floatResult, err := strconv.ParseFloat(result.String(), 64)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("charge is %f\n", floatResult)
}
