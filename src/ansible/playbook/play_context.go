package playbook

import ()

var play_context_fields = map[string]FieldAttribute{
}

type PlayContext struct {
  Base
}

func NewPlayContext(play *Play, options interface{}, passwords []string) *PlayContext {
  pc := new(PlayContext)
  return pc
}
