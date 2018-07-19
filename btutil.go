package btutil

import (
	"fmt"
	"image/color"
	"math/rand"
	"sync"

	"gonum.org/v1/plot/plotutil"

	"gonum.org/v1/gonum/integrate/quad"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat/distuv"
)

func PrintMat(name string, a mat.Matrix) {
	if name != "" {
		fmt.Println(name + ":")
	}
	fmt.Printf("%0.4v\n", mat.Formatted(a))
}

// Provides a long list of colors to use in plots.
var Colors []color.Color

func init() {
	rnd := rand.New(rand.NewSource(1))
	for _, v := range plotutil.SoftColors {
		Colors = append(Colors, v)
	}
	for i := 0; i < 100; i++ {
		c := color.RGBA{
			R: uint8(rnd.Intn(256)),
			G: uint8(rnd.Intn(256)),
			B: uint8(rnd.Intn(256)),
			A: 255,
		}
		Colors = append(Colors, c)
	}
}

func MakeMatGridXYZ(x, y []float64, f func(x, y float64) float64, concurrent int) *MatGridXYZ {
	xn := make([]float64, len(x))
	copy(xn, x)
	yn := make([]float64, len(y))
	copy(yn, y)

	m := mat.NewDense(len(y), len(x), nil)
	if concurrent <= 1 {
		for i := 0; i < len(y); i++ {
			for j := 0; j < len(x); j++ {
				v := f(x[j], y[i])
				m.Set(i, j, v)
			}
		}
	} else {
		type idx2 struct {
			i, j int
		}
		c := make(chan idx2)
		var wg sync.WaitGroup
		for i := 0; i < concurrent; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for idx := range c {
					v := f(x[idx.j], y[idx.i])
					m.Set(idx.i, idx.j, v)
				}
			}()
		}
		for i := 0; i < len(y); i++ {
			for j := 0; j < len(x); j++ {
				c <- idx2{i: i, j: j}
			}
		}
		close(c)
		wg.Wait()
	}
	return &MatGridXYZ{
		XLoc: xn,
		YLoc: yn,
		Data: m,
	}
}

// MatGridXYZ uses a matrix to implement GridXYZ with the x data in
// the columns
type MatGridXYZ struct {
	Data mat.Matrix
	XLoc []float64
	YLoc []float64
}

func (m *MatGridXYZ) Dims() (c, r int) {
	r, c = m.Data.Dims()
	if len(m.XLoc) != c {
		panic("bad x size")
	}
	if len(m.YLoc) != r {
		panic("bad x size")
	}
	return c, r
}

func (m *MatGridXYZ) Z(c, r int) float64 {
	return m.Data.At(r, c)
}

func (m *MatGridXYZ) X(c int) float64 {
	return m.XLoc[c]
}

func (m *MatGridXYZ) Y(r int) float64 {
	return m.YLoc[r]
}

// ExpectedValueFixed computes the expected value of a function under the
// probability distribution represented by the Quantilier using the given number
// of evaluations and level of concurrency.
//
// This is a wrapper around gonum/quad.Fixed that transforms an integral of a function
// and a probability density to an integral over the cumulative density.
func ExpectedValueFixed(f func(float64) float64, q distuv.Quantiler, evals int, concurrent int) float64 {
	// Don't use quad because the error is higher for the same number of quadrature points
	//return distuv.ExpectedFixed(f, q, evals, concurrent)
	fnew := func(p float64) float64 {
		x := q.Quantile(p)
		return f(x)
	}
	return quad.Fixed(fnew, 0, 1, evals, nil, concurrent)
}
