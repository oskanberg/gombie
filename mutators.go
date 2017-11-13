package gombie

import (
	"go/ast"
	"go/token"
)

// Mutator represents a module for mutating target code
type Mutator interface {
	// ShouldMutate should return the mutated version of a given node
	// and a boolean representing whether mutation took place
	Mutate(ast.Node) (ast.Node, bool)
}

// NullMutator is a mutator that does nothing
type NullMutator struct{}

// Mutate does nothing to the node
func (NullMutator) Mutate(node ast.Node) (ast.Node, bool) {
	return node, true
}

// BasicMutators is the set of standard mutators provided by this package
type BasicMutators struct{}

// Mutate runs standard mutations
func (BasicMutators) Mutate(node ast.Node) (ast.Node, bool) {
	node, mutated := MutateEqNeq{}.Mutate(node)
	if mutated {
		return node, true
	}

	node, mutated = MutateIncDec{}.Mutate(node)
	if mutated {
		return node, true
	}

	return node, false
}

// MutateEqNeq turns == into != and vise-versa
type MutateEqNeq struct{}

// Mutate turns == into != and vise-versa
func (MutateEqNeq) Mutate(node ast.Node) (ast.Node, bool) {
	n, ok := node.(*ast.BinaryExpr)
	if !ok {
		return node, false
	}

	if n.Op == token.NEQ {
		n.Op = token.EQL
		return n, true
	}

	if n.Op == token.EQL {
		n.Op = token.NEQ
		return n, true
	}

	return n, false
}

// MutateIncDec turns ++ into -- and vise-versa
type MutateIncDec struct{}

// Mutate turns ++ into -- and vise-versa
func (MutateIncDec) Mutate(node ast.Node) (ast.Node, bool) {
	n, ok := node.(*ast.IncDecStmt)
	if !ok {
		return node, false
	}

	if n.Tok == token.INC {
		n.Tok = token.DEC

	} else if n.Tok == token.DEC {
		n.Tok = token.INC
	}

	return n, true
}
