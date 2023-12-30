package main

import (
	"fmt"
	"os"
	"sidecar/sidecar"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Missing required screen width command argument")
		return
	}

	width, err := strconv.Atoi(os.Args[1])
	if err != nil {
		msg := fmt.Sprintf("Failed to parse screen width argument (%s): %s", os.Args[1], err.Error())
		fmt.Println(msg)
		return
	}

	sc := sidecar.NewSidecar(width, 1)
	if err != nil {
		panic(err)
	}

	err = sc.Run()
	if err != nil {
		panic(err)
	}
}
