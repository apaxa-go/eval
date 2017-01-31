package eval

import (
	"fmt"
	"github.com/apaxa-go/helper/goh/constanth"
	"math"
	"testing"
)

const exampleBenchSrc = `exampleString(fmt.Sprint(interface{}(math.MaxInt64/exampleStruct(struct{ A, B int }{3, 5}).Sum()+int(2)-cap(make([]string, 1, 100))))).String().String() + "."`
const exampleBenchResult exampleString = "!!1152921504606846877!!."

func TestDocSimpleExample(t *testing.T) {
	src := "int8(1*(1+2))"
	expr, err := ParseString(src, "")
	if err != nil {
		t.Error(err)
	}
	r, err := expr.EvalToInterface(nil)
	if err != nil {
		t.Fatal(err)
	}
	if testR := int8(1 * (1 + 2)); r != testR {
		t.Errorf("expect %v, got %v", testR, r)
	}
}

type exampleString string

func (s exampleString) String() exampleString { return "!" + s + "!" }

type exampleStruct struct {
	A, B int
}

func (s exampleStruct) Sum() int { return s.A + s.B }

func TestDocComplicatedExample(t *testing.T) {
	c := make(chan int64, 10)
	c <- 2
	c <- 2

	testR := exampleString(fmt.Sprint(interface{}(math.MaxInt64/exampleStruct(struct{ A, B int }{3, 5}).Sum()+int(<-(<-chan int64)(c))-cap(make([]string, 1, 100))))).String().String() + "."
	src := `exampleString(fmt.Sprint(interface{}(math.MaxInt64/exampleStruct(struct{ A, B int }{3, 5}).Sum()+int(<-(<-chan int64)(c))-cap(make([]string, 1, 100))))).String().String() + "."`
	expr, err := ParseString(src, "")
	if err != nil {
		t.Error(err)
	}
	a := Args{
		"exampleString": MakeTypeInterface(exampleString("")),
		"fmt.Sprint":    MakeDataRegularInterface(fmt.Sprint),
		"math.MaxInt64": MakeDataUntypedConst(constanth.MakeUint(math.MaxInt64)),
		"exampleStruct": MakeTypeInterface(exampleStruct{}),
		"c":             MakeDataRegularInterface(c),
	}
	r, err := expr.EvalToInterface(a)
	if err != nil {
		t.Fatal(err)
	}
	if r != testR {
		t.Errorf("expect %v, got %v", testR, r)
	}
	//fmt.Printf("%v %T\n", r, r)
}

func TestDocErrorExample(t *testing.T) {
	src := `exampleString(fmt.Sprint(interface{}(math.MaxInt64/exampleStruct(struct{ A, B int }{3, 5}).Sum()+int(<-(<-chan int64)(c))-cap(make([]string, 1, 100))))).String().String() + "."`
	expr, err := ParseString(src, "")
	if err != nil {
		t.Error(err)
	}
	a := Args{
		"exampleString": MakeTypeInterface(exampleString("")),
		"fmt.Sprint":    MakeDataRegularInterface(fmt.Sprint),
		"math.MaxInt64": MakeDataUntypedConst(constanth.MakeUint(math.MaxInt64)),
		"exampleStruct": MakeTypeInterface(exampleStruct{}),
		// Remove "c" from passed arguments:
		// "c":             MakeDataRegularInterface(c),
	}
	_, err = expr.EvalToInterface(a)
	if err == nil {
		t.Error("expect error")
	}
	//fmt.Println(err)
}

func BenchmarkDocParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ParseString(exampleBenchSrc, "")
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkDocEval(b *testing.B) {
	expr, err := ParseString(exampleBenchSrc, "")
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		a := Args{
			"exampleString": MakeTypeInterface(exampleString("")),
			"fmt.Sprint":    MakeDataRegularInterface(fmt.Sprint),
			"math.MaxInt64": MakeDataUntypedConst(constanth.MakeUint(math.MaxInt64)),
			"exampleStruct": MakeTypeInterface(exampleStruct{}),
		}
		r, err := expr.EvalToInterface(a)
		if err != nil {
			b.Fatal(err)
		}
		if r != exampleBenchResult {
			b.Fatal("expect %v, got %v", exampleBenchResult, r)
		}
	}
}

func BenchmarkDocGoEval(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := exampleString(fmt.Sprint(interface{}(math.MaxInt64/exampleStruct(struct{ A, B int }{3, 5}).Sum()+int(2)-cap(make([]string, 1, 100))))).String().String() + "."
		if r != exampleBenchResult {
			b.Fatal("expect %v, got %v", exampleBenchResult, r)
		}
	}
}
