package playbook

import (
  "github.com/smallfish/simpleyaml"
  "io/ioutil"
  "path/filepath"
)

type Playbook struct {
  Entries []Play
  BaseDir string
  FileName string
}

func (pb *Playbook) Load(file_name string) {
  pb.FileName = file_name
  cwd, err := filepath.Abs("./")
  if err != nil {
    // FIXME error handling
  } else {
    if filepath.IsAbs(file_name) {
      basedir, _ := filepath.Split(file_name)
      pb.BaseDir = basedir
    } else {
      pb.BaseDir = filepath.Join(cwd, file_name)
    }
    // load any modules relative to the playbook base dir
    EnumerateModules(pb.BaseDir)
  }

  yamlFile, err := ioutil.ReadFile(file_name)
  playbook_data, err := simpleyaml.NewYaml(yamlFile)
  if err != nil {
    // FIXME: error handling
  }

  plays, err := playbook_data.Array()
  if err != nil {
    // FIXME: error handling
  }

  for play_idx := 0; play_idx < len(plays); play_idx++ {
    play_data, _ := playbook_data.GetIndex(play_idx).Map()
    // FIXME: detect include/import_playbook here
    p := NewPlay(play_data)
    pb.Entries = append(pb.Entries, *p)
  }
}

func NewPlaybook(file_name string) *Playbook {
  pb := new(Playbook)
  pb.Load(file_name)
  return pb
}
