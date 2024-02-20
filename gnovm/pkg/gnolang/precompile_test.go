package gnolang

import (
	"go/ast"
	"strings"
	"testing"

	"github.com/jaekwon/testify/assert"
	"github.com/jaekwon/testify/require"
)

func TestPrecompile(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		tags            string
		source          string
		expectedOutput  string
		expectedImports []*ast.ImportSpec
		expectedError   string
	}{
		{
			name: "hello",
			source: `
package foo

func hello() string { return "world"}
`,
			expectedOutput: `
// Code generated by github.com/gnolang/gno. DO NOT EDIT.

package foo

func hello() string { return "world" }
`,
		},
		{
			name: "hello with tags",
			tags: "gno",
			source: `
package foo

func hello() string { return "world"}
`,
			expectedOutput: `
// Code generated by github.com/gnolang/gno. DO NOT EDIT.

//go:build gno

package foo

func hello() string { return "world" }
`,
		},
		{
			name: "use-std",
			source: `
package foo

import "std"

func hello() string {
	_ = std.Foo
	return "world"
}
`,
			expectedOutput: `
// Code generated by github.com/gnolang/gno. DO NOT EDIT.

package foo

import "github.com/gnolang/gno/gnovm/stdlibs/stdshim"

func hello() string {
	_ = std.Foo
	return "world"
}
`,
			expectedImports: []*ast.ImportSpec{
				{
					Path: &ast.BasicLit{
						ValuePos: 21,
						Kind:     9,
						Value:    `"github.com/gnolang/gno/gnovm/stdlibs/stdshim"`,
					},
					EndPos: 26,
				},
			},
		},
		{
			name: "use-realm",
			source: `
package foo

import "gno.land/r/users"

func foo()  { _ = users.Register}
`,
			expectedOutput: `
// Code generated by github.com/gnolang/gno. DO NOT EDIT.

package foo

import "github.com/gnolang/gno/examples/gno.land/r/users"

func foo() { _ = users.Register }
`,
			expectedImports: []*ast.ImportSpec{
				{
					Path: &ast.BasicLit{
						ValuePos: 21,
						Kind:     9,
						Value:    `"github.com/gnolang/gno/examples/gno.land/r/users"`,
					},
					EndPos: 39,
				},
			},
		},
		{
			name: "use-avl",
			source: `
package foo

import "gno.land/p/demo/avl"

func foo()  { _ = avl.Tree }
`,
			expectedOutput: `
// Code generated by github.com/gnolang/gno. DO NOT EDIT.

package foo

import "github.com/gnolang/gno/examples/gno.land/p/demo/avl"

func foo() { _ = avl.Tree }
`,
			expectedImports: []*ast.ImportSpec{
				{
					Path: &ast.BasicLit{
						ValuePos: 21,
						Kind:     9,
						Value:    `"github.com/gnolang/gno/examples/gno.land/p/demo/avl"`,
					},
					EndPos: 42,
				},
			},
		},
		{
			name: "use-named-std",
			source: `
package foo

import bar "std"

func hello() string {
	_ = bar.Foo
	return "world"
}
`,
			expectedOutput: `
// Code generated by github.com/gnolang/gno. DO NOT EDIT.

package foo

import bar "github.com/gnolang/gno/gnovm/stdlibs/stdshim"

func hello() string {
	_ = bar.Foo
	return "world"
}
`,
			expectedImports: []*ast.ImportSpec{
				{
					Name: &ast.Ident{
						NamePos: 21,
						Name:    "bar",
					},
					Path: &ast.BasicLit{
						ValuePos: 25,
						Kind:     9,
						Value:    `"github.com/gnolang/gno/gnovm/stdlibs/stdshim"`,
					},
					EndPos: 30,
				},
			},
		},
		{
			name: "blacklisted-package",
			source: `
package foo

import "reflect"

func foo() { _ = reflect.ValueOf }
`,
			expectedError: `import "reflect" is not in the whitelist`,
		},
		{
			name: "syntax-error",
			source: `
package foo

invalid
`,
			expectedError: `parse: foo.gno:3:1: expected declaration, found invalid`,
		},
		{
			name: "unknown-realm",
			source: `
package foo

import "gno.land/p/demo/unknownxyz"
`,
			expectedOutput: `
// Code generated by github.com/gnolang/gno. DO NOT EDIT.

package foo

import "github.com/gnolang/gno/examples/gno.land/p/demo/unknownxyz"
`,
			expectedImports: []*ast.ImportSpec{
				{
					Path: &ast.BasicLit{
						ValuePos: 21,
						Kind:     9,
						Value:    `"github.com/gnolang/gno/examples/gno.land/p/demo/unknownxyz"`,
					},
					EndPos: 49,
				},
			},
		},
		{
			name: "whitelisted-package",
			source: `
package foo

import "regexp"

func foo() { _ = regexp.MatchString }
`,
			expectedOutput: `
// Code generated by github.com/gnolang/gno. DO NOT EDIT.

package foo

import "regexp"

func foo() { _ = regexp.MatchString }
`,
			expectedImports: []*ast.ImportSpec{
				{
					Path: &ast.BasicLit{
						ValuePos: 21,
						Kind:     9,
						Value:    `"regexp"`,
					},
				},
			},
		},
	}
	for _, c := range cases {
		c := c // scopelint
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			// "\n" is added for better test case readability, now trim it
			source := strings.TrimPrefix(c.source, "\n")

			res, err := Precompile(source, c.tags, "foo.gno")

			if c.expectedError != "" {
				require.EqualError(t, err, c.expectedError)
				return
			}

			if c.expectedError != "" {
				require.EqualError(t, err, c.expectedError)
				return
			}
			require.NoError(t, err)
			expectedOutput := strings.TrimPrefix(c.expectedOutput, "\n")
			assert.Equal(t, res.Translated, expectedOutput, "wrong output")
			assert.Equal(t, res.Imports, c.expectedImports, "wrong imports")
		})
	}
}
