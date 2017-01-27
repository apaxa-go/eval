package eval

import (
	"github.com/apaxa-go/helper/reflecth"
	"reflect"
	"testing"
)

func TestCompositeLitStructKeys(t *testing.T) {
	if r, err := compositeLitStructKeys(reflecth.TypeBool(), nil); r != nil || err == nil {
		t.Errorf("expect %v %v, got %v %v", nil, true, r, err)
	}
}

func TestCompositeLitStructOrdered(t *testing.T) {
	if r, err := compositeLitStructOrdered(reflecth.TypeBool(), nil); r != nil || err == nil {
		t.Errorf("expect %v %v, got %v %v", nil, true, r, err)
	}
}

func TestCompositeLitArrayLike(t *testing.T) {
	//
	//	1
	//
	if r, err := compositeLitArrayLike(reflecth.TypeBool(), nil); r != nil || err == nil {
		t.Errorf("expect %v %v, got %v %v", nil, true, r, err)
	}
	//
	//	2
	//
	if r, err := compositeLitArrayLike(reflect.TypeOf([]int{}), map[int]Data{-1: MakeRegularInterface(1)}); r != nil || err == nil {
		t.Errorf("expect %v %v, got %v %v", nil, true, r, err)
	}
}

func TestCompositeLitMap(t *testing.T) {
	if r, err := compositeLitMap(reflecth.TypeBool(), nil); r != nil || err == nil {
		t.Errorf("expect %v %v, got %v %v", nil, true, r, err)
	}
}
