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

func (u *userDefined) Add(t Term, env *Env) error {
	added, err := compile(t, env)
	if err != nil {
		return err
	}
	u.clauses = append(u.clauses, added...)
	return nil
}

type builtin struct{ clauses }

type clauses []clause

func (cs clauses) Call(vm *VM, args []Term, k func(*Env) *Promise, env *Env) *Promise {
	if vm.OnCall == nil {
		vm.OnCall = func(pi ProcedureIndicator, args []Term, env *Env) {}
	}
	if vm.OnExit == nil {
		vm.OnExit = func(pi ProcedureIndicator, args []Term, env *Env) {}
	}
	if vm.OnFail == nil {
		vm.OnFail = func(pi ProcedureIndicator, args []Term, env *Env) {}
	}
	if vm.OnRedo == nil {
		vm.OnRedo = func(pi ProcedureIndicator, args []Term, env *Env) {}
	}

	var p *Promise
	ks := make([]func(context.Context) *Promise, len(cs))
	for i := range cs {
		i, c := i, cs[i]
		ks[i] = func(context.Context) *Promise {
			if i == 0 {
				vm.OnCall(c.pi, args, env)
			} else {
				vm.OnRedo(c.pi, args, env)
			}
			vars := make([]Variable, len(c.vars))
			for i := range vars {
				vars[i] = NewVariable()
			}
			return Delay(func(context.Context) *Promise {
				env := env
				return vm.exec(registers{
					pc:   c.bytecode,
					xr:   c.xrTable,
					vars: vars,
					cont: func(env *Env) *Promise {
						vm.OnExit(c.pi, args, env)
						return k(env)
					},
					args:      List(args...),
					astack:    List(),
					env:       env,
					cutParent: p,
				})
			}, func(context.Context) *Promise {
				env := env
				vm.OnFail(c.pi, args, env)
				return Bool(false)
			})
		}
	}
	p = Delay(ks...)
	return p
}

func compile(t Term, env *Env) (clauses, error) {
	t = env.Resolve(t)
	if t, ok := t.(Compound); ok && t.Functor() == ":-" && t.Arity() == 2 {
		var cs clauses
		head, body := t.Arg(0), t.Arg(1)
		iter := AltIterator{Alt: body, Env: env}
		for iter.Next() {
			c, err := compileClause(head, iter.Current(), env)
			if err != nil {
				return nil, TypeError(ValidTypeCallable, body, env)
			}
			c.raw = t
			cs = append(cs, c)
		}
		return cs, nil
	}

	c, err := compileClause(t, nil, env)
	c.raw = env.Simplify(t)
	return []clause{c}, err
}

type clause struct {
	pi       ProcedureIndicator
	raw      Term
	xrTable  []Term
	vars     []Variable
	bytecode bytecode
}

func compileClause(head Term, body Term, env *Env) (clause, error) {
	var c clause
	switch head := env.Resolve(head).(type) {
	case Atom:
		c.pi = ProcedureIndicator{Name: head, Arity: 0}
	case Compound:
		c.pi = ProcedureIndicator{Name: head.Functor(), Arity: Integer(head.Arity())}
		for i := 0; i < head.Arity(); i++ {
			c.compileArg(head.Arg(i), env)
		}
	}
	if body != nil {
		if err := c.compileBody(body, env); err != nil {
			return c, TypeError(ValidTypeCallable, body, env)
		}
	}
	c.bytecode = append(c.bytecode, instruction{opcode: opExit})
	return c, nil
}

func (c *clause) compileBody(body Term, env *Env) error {
	c.bytecode = append(c.bytecode, instruction{opcode: opEnter})
	iter := SeqIterator{Seq: body, Env: env}
	for iter.Next() {
		if err := c.compilePred(iter.Current(), env); err != nil {
			return err
		}
	}
	return nil
}

var errNotCallable = errors.New("not callable")

func (c *clause) compilePred(p Term, env *Env) error {
	switch p := env.Resolve(p).(type) {
	case Variable:
		return c.compilePred(&compound{
			functor: "call",
			args:    []Term{p},
		}, env)
	case Atom:
		switch p {
		case "!":
			c.bytecode = append(c.bytecode, instruction{opcode: opCut})
			return nil
		}
		c.bytecode = append(c.bytecode, instruction{opcode: opCall, operand: c.xrOffset(ProcedureIndicator{Name: p, Arity: 0})})
		return nil
	case Compound:
		for i := 0; i < p.Arity(); i++ {
			c.compileArg(p.Arg(i), env)
		}
		c.bytecode = append(c.bytecode, instruction{opcode: opCall, operand: c.xrOffset(ProcedureIndicator{Name: p.Functor(), Arity: Integer(p.Arity())})})
		return nil
	default:
		return errNotCallable
	}
}

func (c *clause) compileArg(a Term, env *Env) {
	switch a := env.Resolve(a).(type) {
	case Variable:
		c.bytecode = append(c.bytecode, instruction{opcode: opVar, operand: c.varOffset(a)})
	case charList, codeList: // Treat them as if they're atomic.
		c.bytecode = append(c.bytecode, instruction{opcode: opConst, operand: c.xrOffset(a)})
	case list:
		c.bytecode = append(c.bytecode, instruction{opcode: opList, operand: c.xrOffset(Integer(len(a)))})
		for _, arg := range a {
			c.compileArg(arg, env)
		}
		c.bytecode = append(c.bytecode, instruction{opcode: opPop})
	case partial:
		prefix := a.Compound.(list)
		c.bytecode = append(c.bytecode, instruction{opcode: opPartial, operand: c.xrOffset(Integer(len(prefix)))})
		c.compileArg(a.tail, env)
		for _, arg := range prefix {
			c.compileArg(arg, env)
		}
		c.bytecode = append(c.bytecode, instruction{opcode: opPop})
	case Compound:
		c.bytecode = append(c.bytecode, instruction{opcode: opFunctor, operand: c.xrOffset(ProcedureIndicator{Name: a.Functor(), Arity: Integer(a.Arity())})})
		for i := 0; i < a.Arity(); i++ {
			c.compileArg(a.Arg(i), env)
		}
		c.bytecode = append(c.bytecode, instruction{opcode: opPop})
	default:
		c.bytecode = append(c.bytecode, instruction{opcode: opConst, operand: c.xrOffset(a)})
	}
}

func (c *clause) xrOffset(o Term) byte {
	id := ID(o)
	for i, r := range c.xrTable {
		if ID(r) == id {
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
