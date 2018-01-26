package playbook

import ()

var play_context_fields = map[string]FieldAttribute{
}

type PlayContext struct {
  Base

  Only_tags []string
  Skip_tags []string
}

func NewPlayContext(play *Play, options interface{}, passwords []string) *PlayContext {
  pc := new(PlayContext)
  pc.Only_tags = make([]string, 0)
  pc.Skip_tags = make([]string, 0)
  return pc
}
