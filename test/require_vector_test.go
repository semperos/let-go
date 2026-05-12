// Regression tests for require accepting both vm.ArrayVector and *vm.ArrayVector.
//
// Before the fix, (require '[ns :as alias]) would fail with:
//
//	require expected Symbol or Vector, got ArrayVector
//
// because the reader produces a vm.ArrayVector value but the switch only matched
// *vm.ArrayVector (pointer).
package test

import (
	"strings"
	"testing"

	"github.com/nooga/let-go/pkg/compiler"
	"github.com/nooga/let-go/pkg/resolver"
	"github.com/nooga/let-go/pkg/rt"
	"github.com/nooga/let-go/pkg/vm"
)

// eval compiles and evaluates a single let-go expression string in a fresh
// compiler context and returns any error.
func evalInCore(t *testing.T, consts *vm.Consts, src string) error {
	t.Helper()
	ns := rt.NS(rt.NameCoreNS)
	ctx := compiler.NewCompiler(consts, ns)
	_, _, err := ctx.CompileMultiple(strings.NewReader(src))
	return err
}

// TestRequireArrayVector is the regression test for the bug where
// (require '[ns :as alias]) failed because the reader produces a vm.ArrayVector
// (value type) rather than a *vm.ArrayVector (pointer type).
func TestRequireArrayVector(t *testing.T) {
	consts := vm.NewConsts()
	// Wire up a resolver so require can locate and load the 'string' namespace.
	loaderCtx := compiler.NewCompiler(consts, rt.NS(rt.NameCoreNS))
	rt.SetNSLoader(resolver.NewNSResolver(loaderCtx, []string{"."}))

	tests := []struct {
		name string
		src  string
	}{
		{
			name: "require vector with :as alias (ArrayVector)",
			// This is the canonical reproduction: the quote syntax produces an
			// ArrayVector value, which previously fell through to the default
			// error case in the require switch.
			src: "(require '[string :as str])",
		},
		{
			name: "require vector with :refer list (ArrayVector)",
			src:  "(require '[string :refer [join]])",
		},
		{
			name: "require vector with :refer :all (ArrayVector)",
			src:  "(require '[string :refer :all])",
		},
		{
			name: "require bare symbol still works",
			src:  "(require 'string)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := evalInCore(t, consts, tc.src); err != nil {
				t.Errorf("unexpected error evaluating %q: %v", tc.src, err)
			}
		})
	}
}
