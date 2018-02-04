package jinja2

import (
  "errors"
)

type PyType int
const (
  PY_TYPE_UNKNOWN PyType = 0
  PY_TYPE_NONE    PyType = 1
  PY_TYPE_STRING  PyType = 2
  PY_TYPE_INT     PyType = 3
  PY_TYPE_FLOAT   PyType = 4
  PY_TYPE_BOOL    PyType = 5
  PY_TYPE_LIST    PyType = 6
  PY_TYPE_TUPLE   PyType = 7
  PY_TYPE_DICT    PyType = 8
  PY_TYPE_IDENT   PyType = 10
)
func PyTypeToString(v PyType) string {
  switch v {
  case PY_TYPE_UNKNOWN:
    return "unknown"
  case PY_TYPE_NONE:
    return "None"
  case PY_TYPE_STRING:
    return "string"
  case PY_TYPE_INT:
    return "int"
  case PY_TYPE_FLOAT:
    return "float"
  case PY_TYPE_BOOL:
    return "bool"
  case PY_TYPE_LIST:
    return "list"
  case PY_TYPE_TUPLE:
    return "tuple"
  case PY_TYPE_DICT:
    return "dict"
  }
  return ""
}

type VariableType struct {
  Type PyType
  Data interface{}
}
func (self *VariableType) AsBool() (bool, error) {
  switch res := self.Type; res {
  case PY_TYPE_NONE:
    return false, nil
  case PY_TYPE_BOOL:
    return self.Data.(bool), nil
  case PY_TYPE_INT:
    if v, ok := self.Data.(int); ok {
      if v == 0 { return false, nil
      } else { return true, nil
      }
    } else { // FIXME: error handling
    }
  case PY_TYPE_FLOAT:
    if v, ok := self.Data.(int); ok {
      if v == 0.0 { return false, nil
      } else { return true, nil
      }
    } else { // FIXME: error handling
    }
  case PY_TYPE_STRING:
    if v, ok := self.Data.(string); ok {
      if v == "" { return false, nil
      } else { return true, nil
      }
    } else { // FIXME: error handling
    }
  }
  return false, errors.New("unknown varaible type, cannot convert it to a boolean value")
}

type Context struct {
  Variables map[string]VariableType
  Filters map[string]func(...interface{})interface{}
  Tests map[string]func(...interface{})interface{}
}
func NewContext(vars map[string]VariableType) *Context {
  c := new(Context)
  if vars != nil {
    c.Variables = vars
  } else {
    c.Variables = make(map[string]VariableType)
  }
  return c
}

//-------------------------------------------------------------------------------------------------
type ForStatement struct {
  Exprs  *ExprList `"for" @@ `
  Target *TestList `"in" @@`
  //IfStatement *IfStatement `[ @@ ]`
  Recursive bool `[@"recursive"]`
}
//-------------------------------------------------------------------------------------------------
type IfStatement struct {
  Test *Test `"if" @@`
}
func (self *IfStatement) Eval(c *Context) (VariableType, error) {
  res, err := self.Test.Eval(c)
  if err != nil {
    return VariableType{PY_TYPE_UNKNOWN, nil}, err
  }
  if v, err := res.AsBool(); err != nil {
    return VariableType{PY_TYPE_UNKNOWN, nil}, err
  } else {
    return VariableType{PY_TYPE_BOOL, v}, nil
  }
}
//-------------------------------------------------------------------------------------------------
type ElifStatement struct {
  Test *Test `"elif" @@`
}
func (self *ElifStatement) Eval(c *Context) (VariableType, error) {
  res, err := self.Test.Eval(c)
  if err != nil {
    return VariableType{PY_TYPE_UNKNOWN, nil}, err
  }
  if v, err := res.AsBool(); err != nil {
    return VariableType{PY_TYPE_UNKNOWN, nil}, err
  } else {
    return VariableType{PY_TYPE_BOOL, v}, nil
  }
}
//-------------------------------------------------------------------------------------------------
type TestList struct {
  Tests []*Test `@@ { "," @@ }[","]`
}
//-------------------------------------------------------------------------------------------------
type Test struct {
  Or *OrTest `@@`
}
func (self *Test) Eval(c *Context) (VariableType, error) {
  return self.Or.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type OrTest struct {
  Ands []*AndTest `@@ { "or" @@ }`
}
func (self *OrTest) Eval(c *Context) (VariableType, error) {
  if self.Ands != nil {
    res := false
    for _, and := range self.Ands {
      if and_res, err := and.Eval(c); err != nil {
        return VariableType{PY_TYPE_UNKNOWN, nil}, err
      } else {
        if v, ok := and_res.Data.(bool); !ok {
          return VariableType{PY_TYPE_UNKNOWN, nil}, errors.New("error converting 'and' result to a boolean value")
        } else {
          res = res || v
          // if any of the ors are true, the result will be
          // true, so short-circuit the first time we find one
          if res {
            break
          }
        }
      }
    }
    return VariableType{PY_TYPE_BOOL, res}, nil
  } else {
    return VariableType{PY_TYPE_BOOL, false}, nil
  }
}
//-------------------------------------------------------------------------------------------------
type AndTest struct {
  Nots []*NotTest `@@ { "and" @@ }`
}
func (self *AndTest) Eval(c *Context) (VariableType, error) {
  if self.Nots != nil {
    res := true
    for _, not := range self.Nots {
      if not_res, err := not.Eval(c); err != nil {
        return VariableType{PY_TYPE_UNKNOWN, nil}, err
      } else {
        if v, ok := not_res.Data.(bool); !ok {
          return VariableType{PY_TYPE_UNKNOWN, nil}, errors.New("error converting 'not' result to a boolean value")
        } else {
          res = res && v
          // if any of the ands are false, the result will be
          // false, so short-circuit the first time we find one
          if !res {
            break
          }
        }
      }
    }
    return VariableType{PY_TYPE_BOOL, res}, nil
  } else {
    return VariableType{PY_TYPE_BOOL, false}, nil
  }
}
//-------------------------------------------------------------------------------------------------
type NotTest struct {
  Negated    *NotTest    `"not" @@`
  Comparison *Comparison `| @@`
}
func (self *NotTest) Eval(c *Context) (VariableType, error) {
  if self.Negated != nil {
    res, err := self.Negated.Eval(c)
    if err != nil {
      return VariableType{PY_TYPE_UNKNOWN, nil}, err
    } else {
      if v, ok := res.Data.(bool); !ok {
        return VariableType{PY_TYPE_UNKNOWN, nil}, errors.New("error converting 'not' result to a boolean value for negation")
      } else {
        return VariableType{PY_TYPE_BOOL, !v}, nil
      }
    }
  } else if self.Comparison != nil {
    return self.Comparison.Eval(c)
  } else {
    return VariableType{PY_TYPE_UNKNOWN, nil}, errors.New("no negated expression nor a comparison was found")
  }
}
//-------------------------------------------------------------------------------------------------
type Comparison struct {
  Expr *Expr `@@`
  OpExpr []*OpExpr `{ @@ }`
}
func (self *Comparison) Eval(c *Context) (VariableType, error) {
  if self.Expr == nil {
    return VariableType{PY_TYPE_UNKNOWN, nil}, errors.New("expression in comparison is nil")
  }
  l_res, l_err := self.Expr.Eval(c)
  if l_err != nil {
    return VariableType{PY_TYPE_UNKNOWN, nil}, l_err
  }
  defer func() (VariableType, error) {
    return VariableType{PY_TYPE_UNKNOWN, nil}, errors.New("unable to compare data")
  }()
  if self.OpExpr != nil {
    final_result := true
    for _, opexpr := range self.OpExpr {
      r_res, r_err := opexpr.Eval(c)
      if r_err != nil {
        return VariableType{PY_TYPE_UNKNOWN, nil}, r_err
      }
      if l_res.Type != r_res.Type {
        return VariableType{PY_TYPE_UNKNOWN, nil}, errors.New("mismatched types for comparison")
      }
      switch *opexpr.Op {
      case "<":
        switch l_res.Type {
        case PY_TYPE_INT:
          final_result = final_result && l_res.Data.(int64) < r_res.Data.(int64)
        default:
          return VariableType{PY_TYPE_UNKNOWN, nil}, errors.New("invalid comparison op'<' for type '" + PyTypeToString(l_res.Type) + "'")
        }
      case ">":
        switch l_res.Type {
        case PY_TYPE_INT:
          final_result = final_result && l_res.Data.(int64) > r_res.Data.(int64)
        default:
          return VariableType{PY_TYPE_UNKNOWN, nil}, errors.New("invalid comparison op'>' for type '" + PyTypeToString(l_res.Type) + "'")
        }
      case "<=":
        switch l_res.Type {
        case PY_TYPE_INT:
          final_result = final_result && l_res.Data.(int64) <= r_res.Data.(int64)
        default:
          return VariableType{PY_TYPE_UNKNOWN, nil}, errors.New("invalid comparison op'<=' for type '" + PyTypeToString(l_res.Type) + "'")
        }
      case ">=":
        switch l_res.Type {
        case PY_TYPE_INT:
          final_result = final_result && l_res.Data.(int64) >= r_res.Data.(int64)
        default:
          return VariableType{PY_TYPE_UNKNOWN, nil}, errors.New("invalid comparison op'>=' for type '" + PyTypeToString(l_res.Type) + "'")
        }
      case "==":
        switch l_res.Type {
        case PY_TYPE_INT:
          final_result = final_result && l_res.Data.(int64) == r_res.Data.(int64)
        default:
          return VariableType{PY_TYPE_UNKNOWN, nil}, errors.New("invalid comparison op'==' for type '" + PyTypeToString(l_res.Type) + "'")
        }
      case "!=":
        switch l_res.Type {
        case PY_TYPE_INT:
          final_result = final_result && l_res.Data.(int64) != r_res.Data.(int64)
        default:
          return VariableType{PY_TYPE_UNKNOWN, nil}, errors.New("invalid comparison op'==' for type '" + PyTypeToString(l_res.Type) + "'")
        }
      }
      l_res = r_res
    }
    return VariableType{PY_TYPE_BOOL, final_result}, nil
  } else {
    if v, err := l_res.AsBool(); err != nil {
      return VariableType{PY_TYPE_UNKNOWN, nil}, err
    } else {
      return VariableType{PY_TYPE_BOOL, v}, nil
    }
  }
}
//-------------------------------------------------------------------------------------------------
type Expr struct {
  Xor  *XorExpr     `@@`
  Xors []*OpXorExpr `{ @@ }`
}
func (self *Expr) Eval(c *Context) (VariableType, error) {
  // FIXME: iterate over xors
  return self.Xor.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type OpExpr struct {
  Op  *string  `@("<"|">"|"=="|">="|"<="|"<>"|"!="|"in"|"not" "in"|"is"|"is" "not")`
  Xor *XorExpr `@@`
}
func (self *OpExpr) Eval(c *Context) (VariableType, error) {
  return self.Xor.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type XorExpr struct {
  AndExpr  *AndExpr     `@@`
  AndExprs []*OpAndExpr `{ @@ }`
}
func (self *XorExpr) Eval(c *Context) (VariableType, error) {
  // FIXME: iterate over ands
  return self.AndExpr.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type OpXorExpr struct {
  Op      *string  `"^"`
  AndExpr *AndExpr `@@`
}
func (self *OpXorExpr) Eval(c *Context) (VariableType, error) {
  // FIXME: iterate over ands
  return self.AndExpr.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type AndExpr struct {
  ShiftExpr  *ShiftExpr     `@@`
  ShiftExprs []*OpShiftExpr `{ @@ }`
}
func (self *AndExpr) Eval(c *Context) (VariableType, error) {
  // FIXME: iterate over ands
  return self.ShiftExpr.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type OpAndExpr struct {
  Op        *string    `"&"`
  ShiftExpr *ShiftExpr `@@`
}
func (self *OpAndExpr) Eval(c *Context) (VariableType, error) {
  // FIXME: iterate over shifts
  return self.ShiftExpr.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type ShiftExpr struct {
  ArithExpr  *ArithExpr     `@@`
  ArithExprs []*OpArithExpr `{ @@ }`
}
func (self *ShiftExpr) Eval(c *Context) (VariableType, error) {
  // FIXME: iterate over ands
  return self.ArithExpr.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type OpShiftExpr struct {
  Op *string `@("<<"|">>")`
  ArithExpr *ArithExpr `@@`
}
func (self *OpShiftExpr) Eval(c *Context) (VariableType, error) {
  // FIXME: iterate over ands
  return self.ArithExpr.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type ArithExpr struct {
  Term *Term `@@`
  Terms []*OpTerm `{ @@ }`
}
func (self *ArithExpr) Eval(c *Context) (VariableType, error) {
  // FIXME: iterate over terms
  return self.Term.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type OpArithExpr struct {
  Op *string `@("+"|"-")`
  Term *Term `@@`
}
func (self *OpArithExpr) Eval(c *Context) (VariableType, error) {
  // FIXME: iterate over terms
  return self.Term.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type Term struct {
  Factor  *Factor   `@@`
  Factors []*Factor `{ @@ }`
}
func (self *Term) Eval(c *Context) (VariableType, error) {
  // FIXME: iterate over factors
  return self.Factor.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type OpTerm struct {
  Op     *string `@("*"|"@"|"/"|"%"|"//")`
  Factor *Factor `@@`
}
func (self *OpTerm) Eval(c *Context) (VariableType, error) {
  return self.Factor.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type Factor struct {
  ModFactor *ModFactor `  @@`
  Power     *Power     `| @@`
}
func (self *Factor) Eval(c *Context) (VariableType, error) {
  if self.ModFactor != nil {
    return self.ModFactor.Eval(c)
  } else if self.Power != nil {
    return self.Power.Eval(c)
  } else {
    return VariableType{PY_TYPE_UNKNOWN, nil}, errors.New("Neither a modfactor nor a power were found while parsing factor.")
  }
}
//-------------------------------------------------------------------------------------------------
type ModFactor struct {
  Mod    *string `@("+"|"-"|"~")`
  Factor *Factor `@@`
}
func (self *ModFactor) Eval(c *Context) (VariableType, error) {
  // FIXME: take the modifier into account
  return self.Factor.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type Power struct {
  AtomExpr *AtomExpr `@@`
  Factor   *Factor   `[ "**" @@ ]`
  Filters  []*Filter `{ "|" @@ }`
}
func (self *Power) Eval(c *Context) (VariableType, error) {
  atom_res, err := self.AtomExpr.Eval(c)
  /*
  // FIXME: implement powers
  if self.Factor != nil {
    r_atom_type, r_atom_res, r_err := self.Factor.Eval(c)
  }
  */
  if self.Filters != nil {
    for _, filter := range self.Filters {
      filter_name := *filter.Name
      filter_args := make([]interface{}, 0)
      if filter.Args != nil {
        if filter.Args.Arguments != nil {
          for _, arg := range filter.Args.Arguments {
            if arg_val, arg_err := arg.Eval(c); arg_err != nil {
              return VariableType{PY_TYPE_UNKNOWN, nil}, arg_err
            } else {
              filter_args = append(filter_args, arg_val)
            }
          }
        }
      }
      if filter_func, ok := c.Filters[filter_name]; !ok {
        return VariableType{PY_TYPE_UNKNOWN, nil}, errors.New("the filter '" + filter_name + "' was not found.")
      } else {
        //atom_res = filter_func(filter_args...)
        _ = filter_func
      }
    }
  }
  return atom_res, err
}
//-------------------------------------------------------------------------------------------------
type AtomExpr struct {
  Atom *Atom `@@`
  Trailers []*Trailer `{ @@ }`
}
func (self *AtomExpr) Eval(c *Context) (VariableType, error) {
  atom_res, err := self.Atom.Eval(c)
  if self.Trailers != nil {
    for _, t := range self.Trailers {
      if t.Name != nil {
        // this is a sub-key in a dictionary or an attribute on the
        // class, so we set the running value to whichever it is.
        if atom_res.Type == PY_TYPE_DICT {
          if sub_dict, ok := atom_res.Data.(map[string]VariableType); !ok {
            // FIXME: error
          } else {
            if v, ok := sub_dict[*t.Name]; !ok {
              // FIXME: error
            } else {
              atom_res = v
            }
          }
        } else {
          // class/struct?
        }
      } else if t.ArgList != nil {
        // this is a callable, so we need to lookup which
        // method is being called and pass the args to it, then
        // we assign the result to the running value.
      }
    }
  }
  return atom_res, err
}
//-------------------------------------------------------------------------------------------------
type Atom struct {
  Name      *string   `  @Ident`
	Str       *string   `| @String`
	Float     *float64  `| @Float`
	Int       *int64    `| @Int`
	Bool      *string   `| @( "True" | "False" )`
  None      *string   `| @"None"`
}
func (self *Atom) Eval(c *Context) (VariableType, error) {
  if self.Name != nil {
    if v, ok := c.Variables[*self.Name]; ok {
      return v, nil
    } else {
      return VariableType{PY_TYPE_UNKNOWN, nil}, errors.New("variable name '"+(*self.Name)+"' was not found in the current context.")
    }
  } else if self.Str != nil { return VariableType{PY_TYPE_STRING, self.Str}, nil
  } else if self.Float != nil { return VariableType{PY_TYPE_FLOAT, self.Float}, nil
  } else if self.Int != nil { return VariableType{PY_TYPE_INT, self.Int}, nil
  } else if self.Bool != nil { return VariableType{PY_TYPE_BOOL, self.Bool}, nil
  } else if self.None != nil { return VariableType{PY_TYPE_NONE, self.None}, nil
  } else { return VariableType{PY_TYPE_UNKNOWN, nil}, errors.New("atomic value was not set")
  }
}
//-------------------------------------------------------------------------------------------------
type Filter struct {
  Name *string  `@Ident`
  Args *ArgList `{ "(" @@ ")" }`
}
//-------------------------------------------------------------------------------------------------
type Trailer struct {
  ArgList *ArgList `  "(" @@ ")"`
  Name    *string  `| "." @Ident`
}
//-------------------------------------------------------------------------------------------------
type ArgList struct {
  Arguments []*Argument `{@@ { "," @@ }[","]}`
}
//-------------------------------------------------------------------------------------------------
type Argument struct {
  Test *Test `@@`
}
func (self *Argument) Eval(c *Context) (VariableType, error) {
  return self.Test.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type ExprList struct {
  Exprs []*Expr `@@ {"," @@ }[","]`
}
//-------------------------------------------------------------------------------------------------
type List struct {
  Items []*Atom `"[" { @@ [ "," ] } "]"`
}
//-------------------------------------------------------------------------------------------------
type Tuple struct {
  Items []*Atom `{ "(" @@ [ "," ] ")" }`
}
//-------------------------------------------------------------------------------------------------
type Map struct {
	Map []*MapItem `| "{" { @@ [ "," ] } "}"`
}
//-------------------------------------------------------------------------------------------------
type MapItem struct {
	Key   *Atom `@@ ":"`
	Value *Atom `@@`
}
//-------------------------------------------------------------------------------------------------
