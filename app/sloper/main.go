package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ButterHost69/sloper/internal/runtime"
)



func start(){
	// Ensure that .sloper exists
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	runtime := runtime.NewRuntime(cwd)	
	runtime.Start(context.Background())
}

func main() {
	start()
}