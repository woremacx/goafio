package main

import (
	"fmt"
	"io"
	"os"

	"github.com/woremacx/goafio"
)

func main() {
	readFile, err := os.Open("work.afz")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer readFile.Close()

	reader := afio.NewReader(readFile)
	for {
		n, err := reader.Next()
		if err == io.EOF {

		}
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("n:%d\n", n)
	}
}
