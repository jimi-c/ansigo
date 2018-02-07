package jinja2

/*
TODO:
* Remove FIXMEs
* testing testing testing
* Implement lists/tuples and maps/sets (and any other builtin types missing from Atom)
* Implement "is [not] in" and "is [not]" comparison checks
* Implement callables
* Once we have callables, implement basic python builtin type methods?
* Cleanup:
  - Move VariableType stuff to a new file
  - Move Context stuff to a new file
  - Hide details of conversion between go types and VariableType stuff by adding helpers
    in Context to load a map of variables and convert them to VariableType. Also only have
    jinja2 output be a string when parsing. Direct use of AST stuff will still return a
    VariableType
* Filters & Tests:
  - Implement default jinja2 filters and tests
  - Make sure that all filters & tests work with parameters being passed in
  - Implement a way to load additional filters/tests via plugins
*/

import (
  "errors"
  "strconv"
)

type PyType int
const (
  PY_TYPE_UNDEFINED PyType = 0
  PY_TYPE_NONE      PyType = 1
  PY_TYPE_STRING    PyType = 2
  PY_TYPE_INT       PyType = 3
  PY_TYPE_FLOAT     PyType = 4
  PY_TYPE_BOOL      PyType = 5
  PY_TYPE_LIST      PyType = 6
  PY_TYPE_TUPLE     PyType = 7
  PY_TYPE_DICT      PyType = 8
  PY_TYPE_IDENT     PyType = 10
)
func PyTypeToString(v PyType) string {
  switch v {
  case PY_TYPE_UNDEFINED:
    return "UNDEFINED"
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
    if v, ok := self.Data.(int64); ok {
      if v == 0 { return false, nil
      } else { return true, nil
      }
    } else { // FIXME: error handling
    }
  case PY_TYPE_FLOAT:
    if v, ok := self.Data.(float64); ok {
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
func (self *VariableType) AsInt() (int64, error) {
  switch res := self.Type; res {
  case PY_TYPE_INT:
    if v, ok := self.Data.(int64); ok {
      return v, nil
    } else {
      return int64(0), errors.New("could not convert variable to a string")
    }
  case PY_TYPE_FLOAT:
    if v, ok := self.Data.(float64); ok {
      return int64(v), nil
    } else {
      return int64(0), errors.New("could not convert variable to a string")
    }
  case PY_TYPE_BOOL:
    if v, ok := self.Data.(bool); ok {
      if v {
        return int64(1), nil
      } else {
        return int64(0), nil
      }
    } else {
      return int64(0), errors.New("could not convert variable to a string")
    }
  default:
    return int64(0), errors.New("could not convert variable to a string")
  }
}
func (self *VariableType) AsFloat() (float64, error) {
  switch res := self.Type; res {
  case PY_TYPE_INT:
    if v, ok := self.Data.(int64); ok {
      return float64(v), nil
    } else {
      return float64(0), errors.New("could not convert variable to a string")
    }
  case PY_TYPE_FLOAT:
    if v, ok := self.Data.(float64); ok {
      return v, nil
    } else {
      return float64(0), errors.New("could not convert variable to a string")
    }
  case PY_TYPE_BOOL:
    if v, ok := self.Data.(bool); ok {
      if v {
        return float64(1.0), nil
      } else {
        return float64(0.0), nil
      }
    } else {
      return float64(0.0), errors.New("could not convert variable to a string")
    }
  default:
    return float64(0.0), errors.New("could not convert variable to a string")
  }
}
func (self *VariableType) AsString() (string, error) {
  switch res := self.Type; res {
  case PY_TYPE_STRING:
    if v, ok := self.Data.(string); ok {
      return v, nil
    } else {
      return "", errors.New("could not convert variable to a string")
    }
  default:
    return "", errors.New("could not convert variable to a string")
  }
}

//-------------------------------------------------------------------------------------------------

type Context struct {
  Variables map[string]VariableType
  Filters map[string]func(VariableType, ...interface{})(VariableType, error)
  Tests map[string]func(VariableType, ...interface{})(VariableType, error)
}
func (self *Context) LoadDefaultFilters() {
  self.Filters["int"] = func(val VariableType, args...interface{}) (VariableType, error) {
    switch val.Type {
    case PY_TYPE_INT:
      return val, nil
    case PY_TYPE_STRING:
      new_v, err := strconv.ParseInt(val.Data.(string), 10, 10)
      if err != nil {
        return VariableType{PY_TYPE_UNDEFINED, nil}, err
      }
      return VariableType{PY_TYPE_INT, new_v}, nil
    }
    return val, nil
  }
  self.Filters["bool"] = func(val VariableType, args...interface{}) (VariableType, error) {
    b_val, err := val.AsBool()
    if err != nil {
      return VariableType{PY_TYPE_UNDEFINED, nil}, err
    } else {
      return VariableType{PY_TYPE_BOOL, b_val}, nil
    }
  }
}
func (self *Context) LoadDefaultTests() {
  self.Tests["defined"] = func(val VariableType, args...interface{}) (VariableType, error) {
    // FIXME: the jinja2 builtin accepts a value as an arg,
    //        which is returned if the test evaluates to true
    if val.Type == PY_TYPE_UNDEFINED {
      return VariableType{PY_TYPE_BOOL, false}, nil
    } else {
      return VariableType{PY_TYPE_BOOL, true}, nil
    }
  }
}
func NewContext(vars map[string]VariableType) *Context {
  c := new(Context)
  if vars != nil {
    c.Variables = vars
  } else {
    c.Variables = make(map[string]VariableType)
  }
  c.Filters = make(map[string]func(VariableType, ...interface{})(VariableType, error))
  c.Tests = make(map[string]func(VariableType, ...interface{})(VariableType, error))
  c.LoadDefaultFilters()
  c.LoadDefaultTests()
  return c
}

//-------------------------------------------------------------------------------------------------
func ProcessFilters(val VariableType, filters []*Filter, c *Context) (VariableType, error) {
  running_res := val
  for _, filter := range filters {
    filter_name := *filter.Name
    filter_args := make([]interface{}, 0)
    if filter.Args != nil {
      if filter.Args.Arguments != nil {
        for _, arg := range filter.Args.Arguments {
          if arg_val, arg_err := arg.Eval(c); arg_err != nil {
            return VariableType{PY_TYPE_UNDEFINED, nil}, arg_err
          } else {
            filter_args = append(filter_args, arg_val)
          }
        }
      }
    }
    if filter_func, ok := c.Filters[filter_name]; !ok {
      return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("the filter '" + filter_name + "' was not found.")
    } else {
      new_res, err := filter_func(running_res, filter_args...)
      if err != nil {
        return VariableType{PY_TYPE_UNDEFINED, nil}, err
      }
      running_res = new_res
    }
  }
  return running_res, nil
}
//-------------------------------------------------------------------------------------------------
func ProcessJ2Test(val VariableType, test *J2Test, c *Context) (VariableType, error) {
  test_name := *test.Name
  test_args := make([]interface{}, 0)
  if test.Args != nil {
    if test.Args.Arguments != nil {
      for _, arg := range test.Args.Arguments {
        if arg_val, arg_err := arg.Eval(c); arg_err != nil {
          return VariableType{PY_TYPE_UNDEFINED, nil}, arg_err
        } else {
          test_args = append(test_args, arg_val)
        }
      }
    }
  }
  if test_func, ok := c.Tests[test_name]; !ok {
    return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("the test '" + test_name + "' was not found.")
  } else {
    new_res, err := test_func(val, test_args...)
    if err != nil {
      return VariableType{PY_TYPE_UNDEFINED, nil}, err
    }
    if *test.Negated == "not" {
      b_val, err := new_res.AsBool()
      if err != nil {
        return VariableType{PY_TYPE_UNDEFINED, nil}, err
      }
      return VariableType{PY_TYPE_BOOL, !b_val}, nil
    }
    // FIXME: should all tests be bools?
    return new_res, nil
  }
}
//-------------------------------------------------------------------------------------------------
type VariableStatement struct {
  Test *Test `@@`
}
func (self *VariableStatement) Eval(c *Context) (VariableType, error) {
  return self.Test.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type ForStatement struct {
  TargetList *TargetList `"for" @@ `
  TestList   *TestList   `"in" @@`
  IfStatement *IfStatement `[ @@ ]`
  Recursive bool `[@"recursive"]`
}
//-------------------------------------------------------------------------------------------------
type IfStatement struct {
  Test *Test `"if" @@`
}
func (self *IfStatement) Eval(c *Context) (VariableType, error) {
  res, err := self.Test.Eval(c)
  if err != nil {
    return VariableType{PY_TYPE_UNDEFINED, nil}, err
  }
  if v, err := res.AsBool(); err != nil {
    return VariableType{PY_TYPE_UNDEFINED, nil}, err
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
    return VariableType{PY_TYPE_UNDEFINED, nil}, err
  }
  if v, err := res.AsBool(); err != nil {
    return VariableType{PY_TYPE_UNDEFINED, nil}, err
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
  Or *OrTest ` @@ `
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
    var cur_res VariableType
    for idx, and := range self.Ands {
      if and_res, err := and.Eval(c); err != nil {
        return VariableType{PY_TYPE_UNDEFINED, nil}, err
      } else {
        if idx == 0 {
          cur_res = and_res
        } else {
          res_bool, err := cur_res.AsBool()
          if err != nil {
            return VariableType{PY_TYPE_UNDEFINED, nil}, err
          }
          and_bool, err := and_res.AsBool()
          if err != nil {
            return VariableType{PY_TYPE_UNDEFINED, nil}, err
          }
          or_result := res_bool || and_bool
          cur_res = VariableType{PY_TYPE_BOOL, or_result}
          // if any of the ors are true, the result will be
          // true, so short-circuit the first time we find one
          if or_result {
            break
          }
        }
      }
    }
    return cur_res, nil
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
    var cur_res VariableType
    for idx, not := range self.Nots {
      if not_res, err := not.Eval(c); err != nil {
        return VariableType{PY_TYPE_UNDEFINED, nil}, err
      } else {
        if idx == 0 {
          cur_res = not_res
        } else {
          res_bool, err := cur_res.AsBool()
          if err != nil {
            return VariableType{PY_TYPE_UNDEFINED, nil}, err
          }
          not_bool, err := not_res.AsBool()
          if err != nil {
            return VariableType{PY_TYPE_UNDEFINED, nil}, err
          }
          and_result := res_bool && not_bool
          cur_res = VariableType{PY_TYPE_BOOL, and_result}
          // if any of the ands are false, the result will be
          // false, so short-circuit the first time we find one
          if !and_result {
            break
          }
        }
      }
    }
    return cur_res, nil
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
      return VariableType{PY_TYPE_UNDEFINED, nil}, err
    } else {
      if v, ok := res.Data.(bool); !ok {
        return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("error converting 'not' result to a boolean value for negation")
      } else {
        return VariableType{PY_TYPE_BOOL, !v}, nil
      }
    }
  } else if self.Comparison != nil {
    return self.Comparison.Eval(c)
  } else {
    return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("no negated expression nor a comparison was found")
  }
}
//-------------------------------------------------------------------------------------------------
type Comparison struct {
  Expr *Expr `@@`
  OpExpr []*OpExpr `{ @@ }`
}
func (self *Comparison) Eval(c *Context) (VariableType, error) {
  if self.Expr == nil {
    return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("expression in comparison is nil")
  }
  l_res, l_err := self.Expr.Eval(c)
  if l_err != nil {
    return VariableType{PY_TYPE_UNDEFINED, nil}, l_err
  }
  defer func() (VariableType, error) {
    return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("unable to compare data")
  }()
  if self.OpExpr != nil {
    final_result := true
    for _, opexpr := range self.OpExpr {
      r_res, r_err := opexpr.Eval(c)
      if r_err != nil {
        return VariableType{PY_TYPE_UNDEFINED, nil}, r_err
      }
      if l_res.Type != r_res.Type {
        return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("mismatched types for comparison")
      }
      switch *opexpr.Op {
      case "<":
        switch l_res.Type {
        case PY_TYPE_INT:
          final_result = final_result && l_res.Data.(int64) < r_res.Data.(int64)
        default:
          return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("invalid comparison op'<' for type '" + PyTypeToString(l_res.Type) + "'")
        }
      case ">":
        switch l_res.Type {
        case PY_TYPE_INT:
          l_val := l_res.Data.(int64)
          r_val := r_res.Data.(int64)
          final_result = final_result && l_val > r_val
        default:
          return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("invalid comparison op'>' for type '" + PyTypeToString(l_res.Type) + "'")
        }
      case "<=":
        switch l_res.Type {
        case PY_TYPE_INT:
          final_result = final_result && l_res.Data.(int64) <= r_res.Data.(int64)
        default:
          return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("invalid comparison op'<=' for type '" + PyTypeToString(l_res.Type) + "'")
        }
      case ">=":
        switch l_res.Type {
        case PY_TYPE_INT:
          final_result = final_result && l_res.Data.(int64) >= r_res.Data.(int64)
        default:
          return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("invalid comparison op'>=' for type '" + PyTypeToString(l_res.Type) + "'")
        }
      case "==":
        switch l_res.Type {
        case PY_TYPE_INT:
          final_result = final_result && l_res.Data.(int64) == r_res.Data.(int64)
        default:
          return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("invalid comparison op'==' for type '" + PyTypeToString(l_res.Type) + "'")
        }
      case "!=":
        switch l_res.Type {
        case PY_TYPE_INT:
          final_result = final_result && l_res.Data.(int64) != r_res.Data.(int64)
        default:
          return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("invalid comparison op'==' for type '" + PyTypeToString(l_res.Type) + "'")
        }
      }
      l_res = r_res
    }
    return VariableType{PY_TYPE_BOOL, final_result}, nil
  } else {
    return l_res, nil
  }
}
//-------------------------------------------------------------------------------------------------
type Expr struct {
  Lhs  *ArithExpr     `@@`
  Rhs []*OpArithExpr `{ @@ }`
}
func (self *Expr) Eval(c *Context) (VariableType, error) {
  l_res, err := self.Lhs.Eval(c)
  if err != nil {
    return VariableType{PY_TYPE_UNDEFINED, nil}, err
  }
  if self.Rhs != nil {
    cur_res := l_res
    for _, rhs := range self.Rhs {
      if rhs_res, err := rhs.Eval(c); err != nil {
        return VariableType{PY_TYPE_UNDEFINED, nil}, err
      } else {
        if cur_res.Type == PY_TYPE_FLOAT || rhs_res.Type == PY_TYPE_FLOAT {
          // need to convert to floats
          l_val, l_err := cur_res.AsFloat()
          r_val, r_err := rhs_res.AsFloat()
          if l_err != nil || r_err != nil {
            return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("unsupported op '"+ (*rhs.Op) +"' for a (fixme) and (fixme))")
          }
          switch *rhs.Op {
          case "+":
            cur_res = VariableType{PY_TYPE_FLOAT, l_val + r_val}
          case "-":
            cur_res = VariableType{PY_TYPE_FLOAT, l_val - r_val}
          }
        } else if cur_res.Type == PY_TYPE_INT || rhs_res.Type == PY_TYPE_INT {
          // convert both to integers
          l_val, l_err := cur_res.AsInt()
          r_val, r_err := rhs_res.AsInt()
          if l_err != nil || r_err != nil {
            return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("unsupported op '"+ (*rhs.Op) +"' for a (fixme) and (fixme))")
          }
          switch *rhs.Op {
          case "+":
            cur_res = VariableType{PY_TYPE_INT, l_val + r_val}
          case "-":
            cur_res = VariableType{PY_TYPE_INT, l_val - r_val}
          }
        } else if cur_res.Type == PY_TYPE_STRING && rhs_res.Type == PY_TYPE_STRING {
          // string join
          if *rhs.Op == "-" {
            return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("unsupported op '"+ (*rhs.Op) +"' for a (string) and (string))")
          }
          l_val, l_err := cur_res.AsString()
          r_val, r_err := rhs_res.AsString()
          if l_err != nil || r_err != nil {
            return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("unsupported op '"+ (*rhs.Op) +"' for a (fixme) and (fixme))")
          }
          cur_res = VariableType{PY_TYPE_STRING, l_val + r_val}
        } else {
          // error, can't do math between disparate types
          return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("unsupported op '"+ (*rhs.Op) +"' for a (fixme) and (fixme))")
        }
      }
    }
    return cur_res, nil
  } else {
    return l_res, nil
  }
}
//-------------------------------------------------------------------------------------------------
type OpExpr struct {
  Op  *string  `@("<"|">"|"=="|">="|"<="|"<>"|"!="|"in"|"is"|"not" "in"|"is" "not")`
  ArithExpr *ArithExpr `@@`
}
func (self *OpExpr) Eval(c *Context) (VariableType, error) {
  return self.ArithExpr.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type ArithExpr struct {
  Lhs *Term `@@`
  Rhs []*OpTerm `{ @@ }`
}
func (self *ArithExpr) Eval(c *Context) (VariableType, error) {
  l_res, err := self.Lhs.Eval(c)
  if err != nil {
    return VariableType{PY_TYPE_UNDEFINED, nil}, err
  }
  if self.Rhs != nil {
    cur_res := l_res
    for _, rhs := range self.Rhs {
      if rhs_res, err := rhs.Eval(c); err != nil {
        return VariableType{PY_TYPE_UNDEFINED, nil}, err
      } else {
        if cur_res.Type == PY_TYPE_FLOAT || rhs_res.Type == PY_TYPE_FLOAT {
          // need to convert to floats
          l_val, l_err := cur_res.AsFloat()
          r_val, r_err := rhs_res.AsFloat()
          if l_err != nil || r_err != nil {
            return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("unsupported op '"+ (*rhs.Op) +"' for a (fixme) and (fixme))")
          }
          switch *rhs.Op {
          case "*":
            cur_res = VariableType{PY_TYPE_FLOAT, l_val * r_val}
          case "/":
            cur_res = VariableType{PY_TYPE_FLOAT, l_val / r_val}
          }
        } else if cur_res.Type == PY_TYPE_INT || rhs_res.Type == PY_TYPE_INT {
          // convert both to integers
          l_val, l_err := cur_res.AsInt()
          r_val, r_err := rhs_res.AsInt()
          if l_err != nil || r_err != nil {
            return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("unsupported op '"+ (*rhs.Op) +"' for a (fixme) and (fixme))")
          }
          switch *rhs.Op {
          case "*":
            cur_res = VariableType{PY_TYPE_INT, l_val * r_val}
          case "/":
            cur_res = VariableType{PY_TYPE_INT, l_val / r_val}
          }
        } else {
          // FIXME: allow `"string" * int` expr
          // error, can't do math between disparate types
          return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("unsupported op '"+ (*rhs.Op) +"' for a (fixme) and (fixme))")
        }
      }
    }
    return cur_res, nil
  } else {
    return l_res, nil
  }
}
//-------------------------------------------------------------------------------------------------
type OpArithExpr struct {
  Op *string `@("+"|"-")`
  Term *Term `@@`
}
func (self *OpArithExpr) Eval(c *Context) (VariableType, error) {
  return self.Term.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type Term struct {
  Factor  *Factor   `@@`
}
func (self *Term) Eval(c *Context) (VariableType, error) {
  return self.Factor.Eval(c)
}
//-------------------------------------------------------------------------------------------------
type OpTerm struct {
  Op     *string `@("*"|"/"|"%"|"//")`
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
    return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("Neither a modfactor nor a power were found while parsing factor.")
  }
}
//-------------------------------------------------------------------------------------------------
type ModFactor struct {
  Mod    *string `@("+"|"-"|"~")`
  Factor *Factor `@@`
}
func (self *ModFactor) Eval(c *Context) (VariableType, error) {
  res, err := self.Factor.Eval(c)
  if err != nil {
    return VariableType{PY_TYPE_UNDEFINED, nil}, err
  }
  switch res.Type {
  case PY_TYPE_INT, PY_TYPE_BOOL:
    i_val, err := res.AsInt()
    if err != nil {
      return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("error while converting variable to an integer")
    } else {
      switch *self.Mod {
      case "+":
        // this doesn't actually do anything in python...
      case "-":
        i_val = -i_val
      case "~":
        i_val = -(i_val+1)
      }
      return VariableType{PY_TYPE_INT, i_val}, nil
    }
  case PY_TYPE_FLOAT:
    f_val, err := res.AsFloat()
    if err != nil {
      return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("error while converting variable to an integer")
    } else {
      switch *self.Mod {
      case "+":
        // this doesn't actually do anything in python...
      case "-":
        f_val = -f_val
      case "~":
        return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("unsupported unary operation '~' on a float")
      }
      return VariableType{PY_TYPE_FLOAT, f_val}, nil
    }
  default:
    return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("unsupported unary operation '"+(*self.Mod)+"' on a FIXME")
  }
}
//-------------------------------------------------------------------------------------------------
type Power struct {
  AtomExpr *AtomExpr `@@`
  Factor   *Factor   `[ "**" @@ ]`
  Filters  []*Filter `{ "|" @@ }`
  Test     *J2Test   `[ @@ ]`
}
func (self *Power) Eval(c *Context) (VariableType, error) {
  atom_res, err := self.AtomExpr.Eval(c)
  if err != nil {
    return VariableType{PY_TYPE_UNDEFINED, nil}, err
  }
  /*
  // FIXME: implement powers
  if self.Factor != nil {
    r_atom_type, r_atom_res, r_err := self.Factor.Eval(c)
  }
  */
  return_res := atom_res
  if self.Filters != nil {
    res, err := ProcessFilters(atom_res, self.Filters, c)
    if err != nil {
      return VariableType{PY_TYPE_UNDEFINED, nil}, err
    }
    return_res = res
  }
  if self.Test != nil {
    res, err := ProcessJ2Test(return_res, self.Test, c)
    if err != nil {
      return VariableType{PY_TYPE_UNDEFINED, nil}, err
    }
    return_res = res
  }
  return return_res, nil
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
	Bool      *string   `| @Bool`
  None      *string   `| @None`
}
func (self *Atom) Eval(c *Context) (VariableType, error) {
  if self.Name != nil {
    if v, ok := c.Variables[*self.Name]; ok {
      return v, nil
    } else {
      return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("variable name '"+(*self.Name)+"' was not found in the current context.")
    }
  } else if self.Str != nil { return VariableType{PY_TYPE_STRING, *self.Str}, nil
  } else if self.Float != nil { return VariableType{PY_TYPE_FLOAT, *self.Float}, nil
  } else if self.Int != nil { return VariableType{PY_TYPE_INT, *self.Int}, nil
  } else if self.Bool != nil {
    if *self.Bool == "True" {
      return VariableType{PY_TYPE_BOOL, true}, nil
    } else {
      return VariableType{PY_TYPE_BOOL, false}, nil
    }
  } else if self.None != nil { return VariableType{PY_TYPE_NONE, nil}, nil
  } else { return VariableType{PY_TYPE_UNDEFINED, nil}, errors.New("atomic value was not set")
  }
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
type TargetList struct {
  Targets []*Target `@@ { "," @@ }[","]`
}
//-------------------------------------------------------------------------------------------------
type Target struct {
  Name *string `@Ident`
}
//-------------------------------------------------------------------------------------------------
type Filter struct {
  Name *string  `@Ident`
  Args *ArgList `{ "(" @@ ")" }`
}
type J2Test struct {
  Negated *string  `"is" @[ "not" ]`
  Name    *string  `@Ident`
  Args    *ArgList `{ "(" @@ ")" }`
}
