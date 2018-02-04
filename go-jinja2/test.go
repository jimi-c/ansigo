package main
import (
  "fmt"
  "os"
  "strconv"
  "./jinja2"
)
func main() {
  input := `{% if foo %}
    Foo was true.
{% elif bar %}
    Bar was true.
{% else %}
    Neither foo nor bar were true.
{% endif %}
`
  c := new(jinja2.Context)
  c.Variables = make(map[string]jinja2.VariableType)
  c.Filters = make(map[string]func(...interface{})interface{})
  c.Tests = make(map[string]func(...interface{})interface{})

  c.Variables["foo"] = jinja2.VariableType{jinja2.PY_TYPE_BOOL, true}
  c.Variables["bar"] = jinja2.VariableType{jinja2.PY_TYPE_BOOL, false}

  for i := 0; i < 1; i++ {
    fmt.Println("PARSING TEMPLATE:")
    fmt.Println(input)
    fmt.Println("------------------------------------------------------------")
    fmt.Println("VARIABLES:")
    fmt.Println(c.Variables)
    fmt.Println("------------------------------------------------------------")
    tokens := jinja2.Tokenize(input)
    //fmt.Println(tokens)
    for pos := 0; pos < len(tokens); {
      new_pos, contained_chunks, err := jinja2.ParseBlocks(tokens, pos, "")
      if err != nil {
        //fmt.Println("ERROR:", err, "at " + strconv.Itoa(tokens[new_pos].GetPos()))
        fmt.Println("ERROR:", err, "at " + strconv.Itoa(new_pos))
        os.Exit(1)
      }
      res := ""
      for _, chunk := range contained_chunks {
        c_res, err := chunk.Render(c)
        if err != nil {
          fmt.Println("ERROR DURING RENDERING OF CHUNK:", err)
          break
        } else {
          res = res + c_res
        }
      }
      fmt.Println("RENDER COMPLETE:")
      fmt.Println(res)
      pos = new_pos
    }
  }
  os.Exit(0)
}
