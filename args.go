package eval

import (
	"errors"
	"github.com/apaxa-go/helper/goh/asth"
	"reflect"
	"strings"
)

type (
	Args  map[string]Value
	ArgsI map[string]interface{}
	ArgsR map[string]reflect.Value
)

func ArgsFromRegulars(x map[string]reflect.Value) Args {
	r := make(Args, len(x))
	for i := range x {
		r[i] = MakeDataRegular(x[i])
	}
	return r
}
func ArgsFromInterfaces(x map[string]interface{}) Args {
	r := make(Args, len(x))
	for i := range x {
		r[i] = MakeDataRegularInterface(x[i])
	}
	return r
}

// Compute package if any
func (args Args) normalize() error {
	packages := make(map[string]Args)

	// Extract args with package specific
	for ident := range args {
		parts := strings.Split(ident, ".")
		switch len(parts) {
		case 1:
			continue
		case 2:
			if parts[0] == "_" || !asth.IsValidIdent(parts[0]) || !asth.IsValidExportedIdent(parts[1]) {
				return errors.New("invalid identifier " + ident)
			}

			if _, ok := packages[parts[0]]; !ok {
				packages[parts[0]] = make(Args)
			}
			packages[parts[0]][parts[1]] = args[ident]
			delete(args, ident)
		default:
			return errors.New("invalid identifier " + ident)
		}
	}

	// Add computed packages
	for pk, pv := range packages {
		// Check for unique package name
		if _, ok := args[pk]; ok {
			return errors.New("something with package name already exists " + pk)
		}
		args[pk] = MakePackage(pv)
	}

	return nil
}

// Make all args addressable
func (args Args) makeAddressable() {
	for ident, arg := range args {
		if arg.Kind() != KindData {
			continue
		}
		if arg.Data().Kind() != Regular {
			continue
		}
		oldV := arg.Data().Regular()
		if oldV.CanAddr() {
			continue
		}

		newV := reflect.New(oldV.Type()).Elem()
		newV.Set(oldV)
		args[ident] = MakeDataRegular(newV)
	}
}

func (args Args) validate() error {
	for ident, arg := range args {
		switch arg.Kind() {
		case Type:
			if arg.Type() == nil {
				return errors.New(ident + ": invalid type nil")
			}
		case KindData:
			if arg.Kind() == KindData && arg.Data().Kind() == Regular && !arg.Data().Regular().IsValid() {
				return errors.New(ident + ": invalid regular data")
			}
		}
	}
	return nil
}
