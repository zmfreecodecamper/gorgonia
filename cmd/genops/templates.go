package main

import (
	"text/template"
)

const arithMetaRaw = `
{{define "TypeDefVV"}}
type {{.Name}} struct { binop }
{{end}}

{{define "TypeDefVS"}}
type {{.Name}}VS struct { binopVS }
{{end}}

{{define "TypeDefSV"}}
type {{.Name}}SV struct { binopSV }
{{end}}

{{- define "Do" -}}
	if err := handleCtx(ctx); err != nil {
		return nil, err
	}

	a := vs[0].(tensor.Tensor)
	b := vs[1].(tensor.Tensor)

	ctx2, task := trace.NewTask(ctx, op.String())
	retVal, err = tensor.{{.Method}}(a, b, tensor.WithContext(ctx2))
	task.End()
	return retVal, err
{{- end -}}
{{- define "PreallocDo" -}}
if err := handleCtx(ctx); err != nil {
		return nil, err
	}

	a := vs[0].(tensor.Tensor)
	b := vs[1].(tensor.Tensor)

	ctx2, task := trace.NewTask(ctx, op.String())
	retVal, err = tensor.{{.Method}}(a, b, tensor.WithReuse(prealloc), tensor.WithContext(ctx2))
	task.End()
	return retVal, err
{{- end -}}

{{/* we don't need to generate Type() for arithmetic Ops */}}
{{define "Type()VV"}}{{end}}
{{define "Type()VS"}}{{end}}
{{define "Type()SV"}}{{end}}
`

const cmpMetaRaw = `
{{define "TypeDefVV"}}
type {{.Name}} struct { binop; retSame bool }
{{end}}

{{define "TypeDefVS"}}
type {{.Name}}VS struct { binopVS; retSame bool }
{{end}}

{{define "TypeDefSV"}}
type {{.Name}}SV struct { binopSV; retSame bool }
{{end}}

{{- define "Do" -}}
	if err := handleCtx(ctx); err != nil {
		return nil, err
	}

	a := vs[0].(tensor.Tensor)
	b := vs[1].(tensor.Tensor)

	// Do the actual operation
	ctx2, task := trace.NewTask(ctx, op.String())
	if op.retSame{
		retVal, err = tensor.{{.Method}}(a, b, tensor.WithContext(ctx2), tensor.AsSameType())
	} else {
		retVal, err = tensor.{{.Method}}(a, b, tensor.WithContext(ctx2))
	}
	task.End()
	return retVal, err
{{- end -}}
{{- define "PreallocDo" -}}
if err := handleCtx(ctx); err != nil {
		return nil, err
	}

	a := vs[0].(tensor.Tensor)
	b := vs[1].(tensor.Tensor)

	ctx2, task := trace.NewTask(ctx, op.String())
	if op.retSame {
	retVal, err = tensor.{{.Method}}(a, b, tensor.WithReuse(prealloc), tensor.WithContext(ctx2), tensor.AsSameType())
	} else {
	retVal, err = tensor.{{.Method}}(a, b, tensor.WithReuse(prealloc), tensor.WithContext(ctx2))
	}
	task.End()
	return retVal, err
{{- end -}}

{{define "Type()VV"}}
// Type returns the type: (·) : a → a → a or (·) :  a → a → b
func (op {{.Name}}) Type() hm.Type{
	a := hm.TypeVariable('a') // (T U) or U
	if op.retSame{
		return hm.NewFnType(a, a, a)
	}
	b := types.MakeDependent(a, tensor.Bool) // (T Bool) or Bool
	return hm.NewFnType(a,a,b)
}
{{end}}
{{define "Type()VS"}}
// Type returns the type: (·) : a → b → a or (·) :  a → b → c
func (op {{.Name}}VS) Type() hm.Type {
	a := hm.TypeVariable('a') // (T U) or U
	b := hm.TypeVariable('b') // U
	if op.retSame{
		return hm.NewFnType(a, b, a)
	}
	c := types.MakeDependent(a, tensor.Bool) // (T Bool) or Bool
	return hm.NewFnType(a,b,c)
}
{{end}}
{{define "Type()SV"}}
// Type returns the type: (·) : a → b → b or (·) :  a → b → c
func (op {{.Name}}SV) Type() hm.Type {
	a := hm.TypeVariable('a') // U
	b := hm.TypeVariable('b') // (T U) or U
	if op.retSame{
		return hm.NewFnType(a, b, b)
	}
	c := types.MakeDependent(b, tensor.Bool) // (T Bool) or Bool
	return hm.NewFnType(a,b,c)
}
{{end}}
`

const binOpRaw = `// {{.Name}} is a tensor-tensor {{.CommentOp}}.
{{- template "TypeDefVV" . -}}

// String implements fmt.Stringer.
func (op {{.Name}}) String() string { return "{{.Symbol}}" }

{{ template "Type()VV" . }}

// Do performs {{.CommentOp}}.
func (op {{.Name}}) Do(ctx context.Context, vs ...values.Value) (retVal values.Value, err error) {
	{{- template "Do" . -}}
}

// PreallocDo performs {{.CommentOp}} but with a preallocated return value.
// PreallocDo allows {{.Name}} to implement ops.PreallocOp.
func (op {{.Name}}) PreallocDo(ctx context.Context, prealloc values.Value, vs ...values.Value) (retVal values.Value, err error) {
	{{- template "PreallocDo" . -}}
}


// {{.Name}}VS is a tensor-scalar {{.CommentOp}}.
{{- template "TypeDefVS" . -}}

// String implements fmt.Stringer.
func (op {{.Name}}VS) String() string { return "{{.Symbol}}·" }

{{ template "Type()VS" . }}

// Do performs {{.CommentOp}}.
func (op {{.Name}}VS) Do(ctx context.Context, vs ...values.Value) (retVal values.Value, err error) {
	{{- template "Do" . -}}
}

// PreallocDo performs {{.CommentOp}} but with a preallocated return value.
// PreallocDo allows {{.Name}}VS to implement ops.PreallocOp.
func (op {{.Name}}VS) PreallocDo(ctx context.Context, prealloc values.Value, vs ...values.Value) (retVal values.Value, err error) {
	{{- template "PreallocDo" . -}}
}


// {{.Name}}SV is a scalar-tensor {{.CommentOp}}.
{{- template "TypeDefSV" . -}}

// String implements fmt.Stringer.
func (op {{.Name}}SV) String() string { return "·{{.Symbol}}" }

{{ template "Type()SV" . }}

// Do performs {{.CommentOp}}.
func (op {{.Name}}SV) Do(ctx context.Context, vs ...values.Value) (retVal values.Value, err error) {
	{{- template "Do" . -}}
}

// PreallocDo performs {{.CommentOp}} but with a preallocated return value.
// PreallocDo allows {{.Name}}SV to implement ops.PreallocOp.
func (op {{.Name}}SV) PreallocDo(ctx context.Context, prealloc values.Value, vs ...values.Value) (retVal values.Value, err error) {
	{{- template "PreallocDo" . -}}
}

`

const binSymDiffRaw = `func (op {{.Name}})SymDiff(inputs []*exprgraph.Node, output *exprgraph.Node, grad *exprgraph.Node) (retVal []*exprgraph.Node, err error){ panic("not implemented" )}

func (op {{.Name}}VS)SymDiff(inputs []*exprgraph.Node, output *exprgraph.Node, grad *exprgraph.Node) (retVal []*exprgraph.Node, err error){ panic("not implemented" )}

func (op {{.Name}}SV)SymDiff(inputs []*exprgraph.Node, output *exprgraph.Node, grad *exprgraph.Node) (retVal []*exprgraph.Node, err error){ panic("not implemented" )}

`

const arithOpTestRaw = `{{ define "varExpected" }}
	var expectedType hm.Type
	var expectedShape shapes.Shape
	var err error
{{end}}
{{define "typeshapecheck"}}
	if expectedType, err = typecheck(op, a, b); err != nil {
		t.Fatalf("Expected {{.}}{} to pass type checking. Err: %v", err)
	}
	if expectedShape, err = shapecheck(op, a, b); err != nil {
		t.Fatalf("Expected {{.}}{} to pass shape checking. Err: %v", err)
	}
{{end}}
{{ define "op.Do"}}
	if c, err = op.Do(context.Background(), a, b); err != nil {
		t.Fatalf("Expected {{.}}{} to work correctly. Err: %v", err)
	}
	assert.Equal(t, expectedType, datatypes.TypeOf(c))
	assert.True(t, expectedShape.Eq(c.Shape()))
{{end}}
{{ define "op.PreallocDo" }}
	c, err = op.PreallocDo(context.Background(), c, a, b)
	if err != nil {
		t.Fatalf("Expected {{.}}{}'s Prealloc to work. Err: %v", err)
	}
	assert.Equal(t, expectedType, datatypes.TypeOf(c))
	assert.True(t, expectedShape.Eq(c.Shape()))
{{ end }}

{{- $VV := ( printf "%v" .Name ) -}}
{{- $VS := ( printf "%vVS" .Name ) -}}
{{- $SV := ( printf "%vSV" .Name ) -}}
func Test{{$VV}}{{if .IsCmpRetTrue}}RetSame{{end}}(t *testing.T) {
	op := {{$VV}}{ {{if .IsCmpRetTrue}}retSame: true{{end}} }
	// basic test
	assert.Equal(t, 2, op.Arity())

	/* Do (using tensor-tensor) */

	// set up
	var a, b, c values.Value
	{{- template "varExpected" }}
	a = {{.AVV}}
	b = {{.BVV}}

	// type and shape checks
	{{-  template "typeshapecheck" $VV }}

	// actually doing and testing
	{{- template "op.Do" $VV -}}
	correct := {{.Correct}}
	assert.Equal(t, correct, c.Data())

	/* PreallocDo (using scalar-scalar to make sure things don't go wrong) */

	// set up
	a = {{.AVV2}}
	b = {{.BVV2}}
	c = {{.CVV}}

	// type and shape checks
	{{- template "typeshapecheck" $VV }}

	// actually PreallocDo-ing and testing
	{{- template "op.PreallocDo" $VV -}}
	correctScalar := {{.CorrectScalar}}
	assert.Equal(t, correctScalar, c.Data())


	// bad cases: fails  typecheck and shapecheck
	a = tensor.New(tensor.WithShape(2, 3), tensor.Of(tensor.Float64))
	b = tensor.New(tensor.WithShape(), tensor.Of(tensor.Float64))
	if expectedType, err = typecheck(op, a, b); err == nil {
		t.Fatalf("Expected {{.Name}}{} to NOT pass type checking. Got ~(%v %v) =  %v ", datatypes.TypeOf(a), datatypes.TypeOf(b), expectedType)
	}
	if expectedShape, err = shapecheck(op, a, b); err == nil {
		t.Fatalf("Expected {{.Name}}{} to NOT pass shape checking. Got expectedShape = %v", expectedShape)
	}

}

func Test{{$VS}}{{if .IsCmpRetTrue}}RetSame{{end}}(t *testing.T) {
	op := {{$VS}}{ {{if .IsCmpRetTrue}}retSame: true{{end}} }
	// basic test
	assert.Equal(t, 2, op.Arity())

	/* Do */

	// set up
	var a, b, c values.Value
	{{- template "varExpected" }}
	a = {{.AVS}}
	b = {{.BVS}}

	// type and shape checks
	{{- template "typeshapecheck" $VS }}

	// actually doing and test
	{{- template "op.Do" $VS -}}
	correct := {{.CorrectVS}}
	assert.Equal(t, correct, c.Data())

	/* PreallocDo */

	// set up - create a new preallocated result
	c = {{.CVS}}

	// actually PreallocDo-ing and checking
	{{- template "op.PreallocDo" $VS -}}
	assert.Equal(t, correct, c.Data())

	/* bad cases: {{$VS}}{} on tensor-tensor */

	b = tensor.New(tensor.WithShape(2, 3), tensor.Of(tensor.Float64))
	// we won't type check because the type system is not a dependent type system, thus
	// {{.Name}}VS : (a → b → a) will always type check without errors
	if expectedShape, err = shapecheck(op, a, b); err == nil {
		t.Fatalf("Expected {{.Name}}{} to NOT pass shape checking. Got %v ~ (%v, %v) = %v", op.ShapeExpr(), a.Shape(), b.Shape(), expectedShape)
	}
}

func Test{{$SV}}{{if .IsCmpRetTrue}}RetSame{{end}}(t *testing.T) {
	op := {{$SV}}{ {{if .IsCmpRetTrue}}retSame: true{{end}}  }
	// basic test
	assert.Equal(t, 2, op.Arity())

	/* Do */

	// set up
	var a, b, c values.Value
	{{- template "varExpected" }}
	a = {{.ASV}}
	b = {{.BSV}}


	// type and shape checks
	{{- template "typeshapecheck" $SV }}

	// actually doing and test
	{{- template "op.Do" $SV -}}
	correct := {{.CorrectSV}}
	assert.Equal(t, correct, c.Data())

	/* PreallocDo */

	// set up - create a new preallocated result
	c = {{.CSV}}

	// actually PreallocDo-ing and checking
	{{- template "op.PreallocDo" $VS -}}
	assert.Equal(t, correct, c.Data())

	/* bad cases: {{.Name}}SV{} on tensor-tensor */

	a = tensor.New(tensor.WithShape(2, 3), tensor.Of(tensor.Float64))
	// we won't type check because the type system is not a dependent type system, thus
	// {{.Name}}SV : (a → b → b) will always type check without errors
	if expectedShape, err = shapecheck(op, a, b); err == nil {
		t.Fatalf("Expected {{.Name}}{} to NOT pass shape checking. Got %v ~ (%v, %v) = %v", op.ShapeExpr(), a.Shape(), b.Shape(), expectedShape)
	}
}

`

var (
	arithMetaTmpl   *template.Template
	arithOpTmpl     *template.Template
	cmpMetaTmpl     *template.Template
	cmpOpTmpl       *template.Template
	binSymDiffTmpl  *template.Template
	arithOpTestTmpl *template.Template
)

func init() {
	arithMetaTmpl = template.Must(template.New("arith meta-templates").Funcs(funcmap).Parse(arithMetaRaw))
	arithOpTmpl = template.Must(arithMetaTmpl.New("arith").Funcs(funcmap).Parse(binOpRaw))
	cmpMetaTmpl = template.Must(template.New("cmp meta-templates").Funcs(funcmap).Parse(cmpMetaRaw))
	cmpOpTmpl = template.Must(cmpMetaTmpl.New("cmp").Funcs(funcmap).Parse(binOpRaw))
	binSymDiffTmpl = template.Must(template.New("binsymdiff").Funcs(funcmap).Parse(binSymDiffRaw))
	arithOpTestTmpl = template.Must(template.New("binopTest").Funcs(funcmap).Parse(arithOpTestRaw))

}