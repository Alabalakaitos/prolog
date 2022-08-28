package engine

import (
	"bytes"
)

// Exception is an error represented by a prolog term.
type Exception struct {
	term Term
}

// NewException creates an exception from a copy of the given term.
func NewException(term Term, env *Env) Exception {
	return Exception{term: renamedCopy(term, nil, env)}
}

// Term returns the underlying term of the exception.
func (e Exception) Term() Term {
	return e.term
}

func (e Exception) Error() string {
	var buf bytes.Buffer
	_ = WriteTerm(&buf, e.term, &defaultWriteOptions, nil)
	return buf.String()
}

// InstantiationError returns an instantiation error exception.
func InstantiationError(env *Env) Exception {
	return NewException(&compound{
		functor: "error",
		args: []Term{
			Atom("instantiation_error"),
			varContext,
		},
	}, env)
}

// ValidType is the correct type for an argument or one of its components.
type ValidType uint8

// ValidType is one of these values.
const (
	ValidTypeAtom ValidType = iota
	ValidTypeAtomic
	ValidTypeByte
	ValidTypeCallable
	ValidTypeCharacter
	ValidTypeCompound
	ValidTypeEvaluable
	ValidTypeInByte
	ValidTypeInCharacter
	ValidTypeInteger
	ValidTypeList
	ValidTypeNumber
	ValidTypePredicateIndicator
	ValidTypePair
	ValidTypeFloat
)

// Term returns an Atom for the ValidType.
func (t ValidType) Term() Term {
	return [...]Atom{
		ValidTypeAtom:               "atom",
		ValidTypeAtomic:             "atomic",
		ValidTypeByte:               "byte",
		ValidTypeCallable:           "callable",
		ValidTypeCharacter:          "character",
		ValidTypeCompound:           "compound",
		ValidTypeEvaluable:          "evaluable",
		ValidTypeInByte:             "in_byte",
		ValidTypeInCharacter:        "in_character",
		ValidTypeInteger:            "integer",
		ValidTypeList:               "list",
		ValidTypeNumber:             "number",
		ValidTypePredicateIndicator: "predicate_indicator",
		ValidTypePair:               "pair",
		ValidTypeFloat:              "float",
	}[t]
}

// TypeError creates a new type error exception.
func TypeError(validType ValidType, culprit Term, env *Env) Exception {
	return NewException(&compound{
		functor: "error",
		args: []Term{
			&compound{
				functor: "type_error",
				args:    []Term{validType.Term(), culprit},
			},
			varContext,
		},
	}, env)
}

// ValidDomain is the domain which the procedure defines.
type ValidDomain uint8

// ValidDomain is one of these values.
const (
	ValidDomainCharacterCodeList ValidDomain = iota
	ValidDomainCloseOption
	ValidDomainFlagValue
	ValidDomainIOMode
	ValidDomainNonEmptyList
	ValidDomainNotLessThanZero
	ValidDomainOperatorPriority
	ValidDomainOperatorSpecifier
	ValidDomainPrologFlag
	ValidDomainReadOption
	ValidDomainSourceSink
	ValidDomainStream
	ValidDomainStreamOption
	ValidDomainStreamOrAlias
	ValidDomainStreamPosition
	ValidDomainStreamProperty
	ValidDomainWriteOption

	ValidDomainOrder
)

// Term returns an Atom for the ValidDomain.
func (vd ValidDomain) Term() Term {
	return [...]Atom{
		ValidDomainCharacterCodeList: "character_code_list",
		ValidDomainCloseOption:       "close_option",
		ValidDomainFlagValue:         "flag_value",
		ValidDomainIOMode:            "io_mode",
		ValidDomainNonEmptyList:      "non_empty_list",
		ValidDomainNotLessThanZero:   "not_less_than_zero",
		ValidDomainOperatorPriority:  "operator_priority",
		ValidDomainOperatorSpecifier: "operator_specifier",
		ValidDomainPrologFlag:        "prolog_flag",
		ValidDomainReadOption:        "read_option",
		ValidDomainSourceSink:        "source_sink",
		ValidDomainStream:            "stream",
		ValidDomainStreamOption:      "stream_option",
		ValidDomainStreamOrAlias:     "stream_or_alias",
		ValidDomainStreamPosition:    "stream_position",
		ValidDomainStreamProperty:    "stream_property",
		ValidDomainWriteOption:       "write_option",
		ValidDomainOrder:             "order",
	}[vd]
}

// DomainError creates a new domain error exception.
func DomainError(validDomain ValidDomain, culprit Term, env *Env) Exception {
	return NewException(&compound{
		functor: "error",
		args: []Term{
			&compound{
				functor: "domain_error",
				args:    []Term{validDomain.Term(), culprit},
			},
			varContext,
		},
	}, env)
}

// ObjectType is the object on which an operation is to be performed.
type ObjectType uint8

// ObjectType is one of these values.
const (
	ObjectTypeProcedure ObjectType = iota
	ObjectTypeSourceSink
	ObjectTypeStream
)

// Term returns an Atom for the ObjectType.
func (ot ObjectType) Term() Term {
	return [...]Atom{
		ObjectTypeProcedure:  "procedure",
		ObjectTypeSourceSink: "source_sink",
		ObjectTypeStream:     "stream",
	}[ot]
}

// ExistenceError creates a new existence error exception.
func ExistenceError(objectType ObjectType, culprit Term, env *Env) Exception {
	return NewException(&compound{
		functor: "error",
		args: []Term{
			&compound{
				functor: "existence_error",
				args:    []Term{objectType.Term(), culprit},
			},
			varContext,
		},
	}, env)
}

// Operation is the operation to be performed.
type Operation uint8

// Operation is one of these values.
const (
	OperationAccess Operation = iota
	OperationCreate
	OperationInput
	OperationModify
	OperationOpen
	OperationOutput
	OperationReposition
)

// Term returns an Atom for the Operation.
func (o Operation) Term() Term {
	return [...]Atom{
		OperationAccess:     "access",
		OperationCreate:     "create",
		OperationInput:      "input",
		OperationModify:     "modify",
		OperationOpen:       "open",
		OperationOutput:     "output",
		OperationReposition: "reposition",
	}[o]
}

// PermissionType is the type to which the operation is not permitted to perform.
type PermissionType uint8

// PermissionType is one of these values.
const (
	PermissionTypeBinaryStream PermissionType = iota
	PermissionTypeFlag
	PermissionTypeOperator
	PermissionTypePastEndOfStream
	PermissionTypePrivateProcedure
	PermissionTypeStaticProcedure
	PermissionTypeSourceSink
	PermissionTypeStream
	PermissionTypeTextStream
)

// Term returns an Atom for the PermissionType.
func (pt PermissionType) Term() Term {
	return [...]Atom{
		PermissionTypeBinaryStream:     "binary_stream",
		PermissionTypeFlag:             "flag",
		PermissionTypeOperator:         "operator",
		PermissionTypePastEndOfStream:  "past_enf_of_stream",
		PermissionTypePrivateProcedure: "private_procedure",
		PermissionTypeStaticProcedure:  "static_procedure",
		PermissionTypeSourceSink:       "source_sink",
		PermissionTypeStream:           "stream",
		PermissionTypeTextStream:       "text_stream",
	}[pt]
}

// PermissionError creates a new permission error exception.
func PermissionError(operation Operation, permissionType PermissionType, culprit Term, env *Env) Exception {
	return NewException(&compound{
		functor: "error",
		args: []Term{
			&compound{
				functor: "permission_error",
				args:    []Term{operation.Term(), permissionType.Term(), culprit},
			},
			varContext,
		},
	}, env)
}

// Flag is an implementation defined limit.
type Flag uint8

// Flag is one of these values.
const (
	FlagCharacter Flag = iota
	FlagCharacterCode
	FlagInCharacterCode
	FlagMaxArity
	FlagMaxInteger
	FlagMinInteger
)

// Term returns an Atom for the Flag.
func (f Flag) Term() Term {
	return [...]Atom{
		FlagCharacter:       "character",
		FlagCharacterCode:   "character_code",
		FlagInCharacterCode: "in_character_code",
		FlagMaxArity:        "max_arity",
		FlagMaxInteger:      "max_integer",
		FlagMinInteger:      "min_integer",
	}[f]
}

// RepresentationError creates a new representation error exception.
func RepresentationError(limit Flag, env *Env) Exception {
	return NewException(&compound{
		functor: "error",
		args: []Term{
			&compound{
				functor: "representation_error",
				args:    []Term{limit.Term()},
			},
			varContext,
		},
	}, env)
}

// Resource is a resource required to complete execution.
type Resource uint8

// Resource is one of these values.
const (
	ResourceFiniteMemory Resource = iota
	resourceLen
)

// Term returns an Atom for the Resource.
func (r Resource) Term() Term {
	return [...]Atom{
		ResourceFiniteMemory: "finite_memory",
	}[r]
}

// ResourceError creates a new resource error exception.
func ResourceError(resource Resource, env *Env) Exception {
	return NewException(&compound{
		functor: "error",
		args: []Term{
			&compound{
				functor: "resource_error",
				args:    []Term{resource.Term()},
			},
			varContext,
		},
	}, env)
}

// SyntaxError creates a new syntax error exception.
func SyntaxError(err error, env *Env) Exception {
	return NewException(&compound{
		functor: "error",
		args: []Term{
			&compound{
				functor: "syntax_error",
				args:    []Term{Atom(err.Error())},
			},
			varContext,
		},
	}, env)
}

// SystemError creates a new system error exception.
func SystemError(err error) Exception {
	return NewException(&compound{
		functor: "error",
		args: []Term{
			Atom("system_error"),
			Atom(err.Error()),
		},
	}, nil)
}

// ExceptionalValue is an evaluable functor's result which is not a number.
type ExceptionalValue uint8

// ExceptionalValue is one of these values.
const (
	ExceptionalValueFloatOverflow ExceptionalValue = iota
	ExceptionalValueIntOverflow
	ExceptionalValueUnderflow
	ExceptionalValueZeroDivisor
	ExceptionalValueUndefined
)

func (ev ExceptionalValue) Error() string {
	return string(ev.Term().(Atom))
}

// Term returns an Atom for the ExceptionalValue.
func (ev ExceptionalValue) Term() Term {
	return [...]Atom{
		ExceptionalValueFloatOverflow: "float_overflow",
		ExceptionalValueIntOverflow:   "int_overflow",
		ExceptionalValueUnderflow:     "underflow",
		ExceptionalValueZeroDivisor:   "zero_divisor",
		ExceptionalValueUndefined:     "undefined",
	}[ev]
}

// EvaluationError creates a new evaluation error exception.
func EvaluationError(ev ExceptionalValue, env *Env) Exception {
	return NewException(&compound{
		functor: "error",
		args: []Term{
			&compound{
				functor: "evaluation_error",
				args:    []Term{ev.Term()},
			},
			varContext,
		},
	}, env)
}
