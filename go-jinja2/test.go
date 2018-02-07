package main
import (
  "fmt"
  "os"
  "strconv"
  "./jinja2"
  //"github.com/alecthomas/participle"
  "encoding/json"
)

func PrettyPrint(v interface{}) {
  b, _ := json.MarshalIndent(v, "", "  ")
  println(string(b))
}

func main() {
  input := `{% for foo in 10,20,30 %}{{loop.index}}=>{{foo}}({% if loop.nextitem is defined %}{{loop.nextitem}}{% else %}no next item{% endif %}) {% endfor %}`
  c := jinja2.NewContext(nil)

  c.Variables["foo"] = jinja2.VariableType{jinja2.PY_TYPE_BOOL, false}
  c.Variables["bar"] = jinja2.VariableType{jinja2.PY_TYPE_BOOL, false}
  c.Variables["bam"] = jinja2.VariableType{jinja2.PY_TYPE_STRING, "2"}

  for i := 0; i < 1; i++ {
    tokens := jinja2.Tokenize(input)
    for pos := 0; pos < len(tokens); {
      new_pos, contained_chunks, err := jinja2.ParseBlocks(tokens, pos, "")
      if err != nil {
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
      _ = res
      fmt.Println("RENDER COMPLETE:")
      fmt.Println(res)
      pos = new_pos
    }
  }

  os.Exit(0)
}
