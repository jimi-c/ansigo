package main
import (
  "fmt"
  "os"
  "./jinja2"
)

/*
TODO:
* Remove FIXMEs
* testing testing and more testing
* Better error handling, with line/col info attached to error
* Implement lists/tuples and maps/sets (and any other builtin types missing from Atom)
* Implement "is [not] in" and "is [not]" comparison checks
* Implement callables
* Once we have callables, implement basic python builtin type methods?
* Cleanup:
  - Hide details of conversion between go types and VariableType stuff by adding helpers
    in Context to load a map of variables and convert them to VariableType. Also only have
    jinja2 output be a string when parsing. Direct use of AST stuff will still return a
    VariableType
* Filters & Tests:
  - Implement default jinja2 filters and tests
  - Make sure that all filters & tests work with parameters being passed in
  - Implement a way to load additional filters/tests via plugins
* Implement other jinja2 constructs (currently only support if/for/raw and variables)
  - includes and blocks next
* Implement left/right stripping of newlines when "{%- -%}" / are used
*/

func main() {
  /*
  input := `{% for foo in 10,20,30 if foo != 10 %}{# comment #}
{{loop.index}}=>{{foo}}({% if loop.nextitem is defined %}{{loop.nextitem}}{% else %}no next item{% endif %})
{% endfor %}`
*/
  input := `
{{1+2}}
{{2-1}}
{{"2"+bam}}
{{42.0/42.33}}
{{ "foo" }}
{{bam|int * 2}}
{{foo / 2}}
{% if True %}Hello world{% else %}Goodbye world{%endif%}
{% if False %}Hello world{% else %}Goodbye world{%endif%}
{%for i in 1, 2, 3 %}{{i}}{%endfor%}
{%for i in 1, 2, 3 if i > 1 %}{{i}}{%endfor%}
`
  c := jinja2.NewContext(nil)

  c.Variables["foo"] = jinja2.VariableType{jinja2.PY_TYPE_BOOL, false}
  c.Variables["bar"] = jinja2.VariableType{jinja2.PY_TYPE_BOOL, false}
  c.Variables["bam"] = jinja2.VariableType{jinja2.PY_TYPE_STRING, "2"}

  template := new(jinja2.Template)
  template.Parse(input)
  for i := 0; i < 1; i++ {
    res, err := template.Render(c)
    if err != nil {
      fmt.Println("Template rendering error:", err)
      os.Exit(1)
    }
    fmt.Println("RENDER COMPLETE:")
    fmt.Println(res)
  }

  os.Exit(0)
}
