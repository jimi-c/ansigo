package plugins

import (
  "os"
  "path/filepath"
)

func GetExecutableDir() string {
  ex, err := os.Executable()
  if err != nil {
      panic(err)
  }
  return filepath.Dir(ex)
}
