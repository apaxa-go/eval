package eval

import (
	"reflect"
	"testing"
)

func TestSlice2(t *testing.T) {
	if r, err := slice2(reflect.ValueOf([]int{1, 2, 3, 4, 5}), -2, 2); r != nil || err == nil {
		t.Errorf("expect %v %v, got %v %v", nil, true, r, err)
	}
}

func TestSlice3(t *testing.T) {
	if r, err := slice3(reflect.ValueOf([]int{1, 2, 3, 4, 5}), 1, -1, 3); r != nil || err == nil {
		t.Errorf("expect %v %v, got %v %v", nil, true, r, err)
	}
	if r, err := slice3(reflect.ValueOf([]int{1, 2, 3, 4, 5}), 1, 2, -1); r != nil || err == nil {
		t.Errorf("expect %v %v, got %v %v", nil, true, r, err)
	}
	if r, err := slice3(reflect.ValueOf([]int{1, 2, 3, 4, 5}), -2, 2, 3); r != nil || err == nil {
		t.Errorf("expect %v %v, got %v %v", nil, true, r, err)
	}
}
