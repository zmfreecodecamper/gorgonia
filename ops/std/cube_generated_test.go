package stdops

import (
	"context"
	"testing"

	"github.com/chewxy/hm"
	"github.com/stretchr/testify/assert"
	"gorgonia.org/gorgonia/internal/datatypes"
	"gorgonia.org/gorgonia/values"
	"gorgonia.org/shapes"
	"gorgonia.org/tensor"
)

// Code generated by genops, which is a ops generation tool for Gorgonia. DO NOT EDIT.

func Test_cube(t *testing.T) {
	op := cubeOp{}
	// basic test
	assert.Equal(t, 1, op.Arity())

	// Do
	var a, b values.Value
	a = tensor.New(tensor.WithShape(2, 3), tensor.WithBacking([]float64{-1, -2, -3, 4, 5, 6}))

	var expectedType hm.Type
	var expectedShape shapes.Shape
	var err error

	if expectedType, err = typecheck(op, a); err != nil {
		t.Fatalf("cubeOp failed to typecheck. Err: %v", err)
	}

	if expectedShape, err = shapecheck(op, a); err != nil {
		t.Fatalf("cubeOp failed to shapecheck. Err: %v", err)
	}

	if b, err = op.Do(context.Background(), a); err != nil {
		t.Fatalf("Expected cubeOp{} to work correctly. Err: %v", err)
	}
	assert.Equal(t, expectedType, datatypes.TypeOf(b))
	assert.True(t, expectedShape.Eq(b.Shape()))
	correct := []float64{-1, -2 * -2 * -2, -3 * -3 * -3, 4 * 4 * 4, 5 * 5 * 5, 6 * 6 * 6}
	assert.Equal(t, correct, b.Data())

	// PreallocDo
	b = tensor.New(tensor.WithShape(2, 3), tensor.WithBacking([]float64{-100, -100, -100, -100, -100, -100}))
	if b, err = op.PreallocDo(context.Background(), b, a); err != nil {
		t.Fatalf("Expected cubeOp{} to work correctly. Err: %v", err)
	}
	assert.Equal(t, expectedType, datatypes.TypeOf(b))
	assert.True(t, expectedShape.Eq(b.Shape()))
	assert.Equal(t, correct, b.Data())
}
