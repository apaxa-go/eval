package eval

import (
	"reflect"
	"testing"
)

func TestIndexMap(t *testing.T) {
	if r, err := indexMap(reflect.ValueOf(1), MakeRegularInterface(1)); r != nil || err == nil {
		t.Errorf("expect %v %v, got %v %v", nil, true, r, err)
	}
}
