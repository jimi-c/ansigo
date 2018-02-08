package jinja2

import (
  "encoding/json"
)

func PrettyPrint(v interface{}) {
  b, _ := json.MarshalIndent(v, "", "  ")
  println(string(b))
}
