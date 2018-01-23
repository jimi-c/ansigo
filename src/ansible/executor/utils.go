package executor

import (
  "strconv"
  "strings"
)

func PctToInt(value interface{}, num_items int, min_value int) int {
  pct := min_value
  if str_value, ok := value.(string); ok && strings.HasSuffix(str_value, "%") {
    str_value = str_value[:len(str_value)-1]
    float_value, _ := strconv.ParseFloat(str_value, 64)
    pct = int((float_value / 100.0) * float64(num_items))
  } else if ok {
    pct, _ = strconv.Atoi(str_value)
  } else {
    pct = value.(int)
  }
  if pct < min_value {
    pct = min_value
  }

  return pct
}
