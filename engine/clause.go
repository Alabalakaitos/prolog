package engine

import (
	"context"
	"errors"
)

type userDefined struct {
	public        bool
	dynamic       bool
	multifile     bool
	discontiguous bool

	// 7.4.3 says "If no clauses are defined for a procedure indicated by a directive ... then the procedure shall exist but have no clauses."
	clauses
}

type clauses []clause

func (cs clauses) call(ctx context.Context, args []Term) *Promise {
	vm, _ := vm(ctx)
	k := cont(ctx)

	var p *Promise
	ks := make([]func() *Promise, len(cs))
	for i := range cs {
		i, c := i, cs[i]
		ks[i] = func() *Promise {
			vars := make([]Variable, len(c.vars))
			for i := range vars {
				vars[i] = NewVariable()
			}
			return vm.exec(ctx, registers{
				pc:        c.bytecode,
				xr:        c.xrTable,
				vars:      vars,
				cont:      k,
				args:      List(args...),
				astack:    List(),
				ctx:       ctx,
				cutParent: p,
			})
		}
	}
	p = Delay(ks...)
	return p
}

func compile(ctx context.Context, t Term) (clauses, error) {
	t = Resolve(ctx, t)
	if t, ok := t.(Compound); ok && t.Functor() == atomIf && t.Arity() == 2 {
		var cs clauses
		head, body := t.Arg(0), t.Arg(1)
		iter := altIterator{Alt: body}
		for iter.Next(ctx) {
			c, err := compileClause(ctx, head, iter.Current())
			if err != nil {
				return nil, typeError(ctx, validTypeCallable, body)
			}
			c.raw = t
			cs = append(cs, c)
		}
		return cs, nil
	}

	c, err := compileClause(ctx, t, nil)
	c.raw = renamedCopy(ctx, t, nil)
	return []clause{c}, err
}

type clause struct {
	pi       procedureIndicator
	raw      Term
	xrTable  []Term
	vars     []Variable
	bytecode bytecode
}

func compileClause(ctx context.Context, head Term, body Term) (clause, error) {
	var c clause
	switch head := Resolve(ctx, head).(type) {
	case Atom:
		c.pi = procedureIndicator{name: head, arity: 0}
	case Compound:
		c.pi = procedureIndicator{name: head.Functor(), arity: Integer(head.Arity())}
		for i := 0; i < head.Arity(); i++ {
			c.compileArg(ctx, head.Arg(i))
		}
	}
	if body != nil {
		if err := c.compileBody(ctx, body); err != nil {
			return c, typeError(ctx, validTypeCallable, body)
		}
	}
	c.bytecode = append(c.bytecode, instruction{opcode: opExit})
	return c, nil
}

func (c *clause) compileBody(ctx context.Context, body Term) error {
	c.bytecode = append(c.bytecode, instruction{opcode: opEnter})
	iter := seqIterator{Seq: body}
	for iter.Next(ctx) {
		if err := c.compilePred(ctx, iter.Current()); err != nil {
			return err
		}
	}
	return nil
}

var errNotCallable = errors.New("not callable")

func (c *clause) compilePred(ctx context.Context, p Term) error {
	switch p := Resolve(ctx, p).(type) {
	case Variable:
		return c.compilePred(ctx, atomCall.Apply(p))
	case Atom:
		switch p {
		case atomCut:
			c.bytecode = append(c.bytecode, instruction{opcode: opCut})
			return nil
		}
		c.bytecode = append(c.bytecode, instruction{opcode: opCall, operand: c.xrOffset(procedureIndicator{name: p, arity: 0})})
		return nil
	case Compound:
		for i := 0; i < p.Arity(); i++ {
			c.compileArg(ctx, p.Arg(i))
		}
		c.bytecode = append(c.bytecode, instruction{opcode: opCall, operand: c.xrOffset(procedureIndicator{name: p.Functor(), arity: Integer(p.Arity())})})
		return nil
	default:
		return errNotCallable
	}
}

func (c *clause) compileArg(ctx context.Context, a Term) {
	switch a := Resolve(ctx, a).(type) {
	case Variable:
		c.bytecode = append(c.bytecode, instruction{opcode: opVar, operand: c.varOffset(a)})
	case charList, codeList: // Treat them as if they're atomic.
		c.bytecode = append(c.bytecode, instruction{opcode: opConst, operand: c.xrOffset(a)})
	case list:
		c.bytecode = append(c.bytecode, instruction{opcode: opList, operand: c.xrOffset(Integer(len(a)))})
		for _, arg := range a {
			c.compileArg(ctx, arg)
		}
		c.bytecode = append(c.bytecode, instruction{opcode: opPop})
	case *partial:
		prefix := a.Compound.(list)
		c.bytecode = append(c.bytecode, instruction{opcode: opPartial, operand: c.xrOffset(Integer(len(prefix)))})
		c.compileArg(ctx, *a.tail)
		for _, arg := range prefix {
			c.compileArg(ctx, arg)
		}
		c.bytecode = append(c.bytecode, instruction{opcode: opPop})
	case Compound:
		c.bytecode = append(c.bytecode, instruction{opcode: opFunctor, operand: c.xrOffset(procedureIndicator{name: a.Functor(), arity: Integer(a.Arity())})})
		for i := 0; i < a.Arity(); i++ {
			c.compileArg(ctx, a.Arg(i))
		}
		c.bytecode = append(c.bytecode, instruction{opcode: opPop})
	default:
		c.bytecode = append(c.bytecode, instruction{opcode: opConst, operand: c.xrOffset(a)})
	}
}

func (c *clause) xrOffset(o Term) byte {
	oid := id(o)
	for i, r := range c.xrTable {
		if id(r) == oid {
			return byte(i)
		}
	}
	c.xrTable = append(c.xrTable, o)
	return byte(len(c.xrTable) - 1)
}

func (c *clause) varOffset(o Variable) byte {
	for i, v := range c.vars {
		if v == o {
			return byte(i)
		}
	}
	c.vars = append(c.vars, o)
	return byte(len(c.vars) - 1)
}
