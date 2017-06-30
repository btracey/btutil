package btutil

import (
	"fmt"

	"gonum.org/v1/gonum/mat"
)

func PrintMat(name string, a mat.Matrix) {
	if name != "" {
		fmt.Println(name + ":")
	}
	fmt.Printf("%0.4v\n", mat.Formatted(a))
}
