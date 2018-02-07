package main
import (
  "fmt"
  "os"
  "strconv"
  "./jinja2"
  "encoding/json"
)

func PrettyPrint(v interface{}) {
  b, _ := json.MarshalIndent(v, "", "  ")
  println(string(b))
}

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
`
  c := jinja2.NewContext(nil)

  c.Variables["foo"] = jinja2.VariableType{jinja2.PY_TYPE_BOOL, false}
  c.Variables["bar"] = jinja2.VariableType{jinja2.PY_TYPE_BOOL, false}
  c.Variables["bam"] = jinja2.VariableType{jinja2.PY_TYPE_STRING, "2"}

  tokens := jinja2.Tokenize(input)
  template_chunks := make([]jinja2.Renderable, 0)
  for pos := 0; pos < len(tokens); {
    new_pos, contained_chunks, err := jinja2.ParseBlocks(tokens, pos, "")
    if err != nil {
      fmt.Println("ERROR:", err, "at " + strconv.Itoa(new_pos))
      os.Exit(1)
    }
    template_chunks = append(template_chunks, contained_chunks...)
    pos = new_pos
  }
  for i := 0; i < 1; i++ {
    res := ""
    for _, chunk := range template_chunks {
      c_res, err := chunk.Render(c)
      if err != nil {
        fmt.Println("ERROR DURING RENDERING OF CHUNK:", err)
        break
      } else {
        res = res + c_res
      }
    }
    _ = res
    fmt.Println("RENDER COMPLETE:")
    fmt.Println(res)
  }

  os.Exit(0)
}
