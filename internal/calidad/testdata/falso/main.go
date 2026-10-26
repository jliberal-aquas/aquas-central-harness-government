// Falso golangci-lint para las pruebas: versión, salida y código de salida vienen del entorno.
package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("golangci-lint has version " + os.Getenv("FALSO_VERSION") + " built with go")
		return
	}
	fmt.Print(os.Getenv("FALSO_SALIDA"))
	c, _ := strconv.Atoi(os.Getenv("FALSO_EXIT"))
	os.Exit(c)
}
