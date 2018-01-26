package playbook

import (
)

type TaggableEvaluate interface {
  Tags() []string
}
var taggable_fields = map[string]FieldAttribute{
  "tags": FieldAttribute{T: "list", Default: nil, Extend: true},
}

type Taggable struct {
  Attr_tags interface{}
}

func (t *Taggable) Load(data map[interface{}]interface{}) {
  LoadValidFields(t, taggable_fields, data)
}

func Disjoint(a []string, b []string) bool {
  for _, av := range a {
    for _, bv := range b {
      if av == bv { return false }
    }
  }
  return true
}

// FIXME: do we need vars when evaluating tags here?
func EvaluateTags(thing TaggableEvaluate, only_tags []string, skip_tags []string) bool {
  should_run := true
  only_untagged := false

  tags := thing.Tags()
  if len(tags) > 0 {
    //templar = Templar(loader=self._loader, variables=all_vars)
    //tags = templar.template(self.tags)
    tag_map := make(map[string]bool)
    tag_list := make([]string, 0)
    for _, v := range tags {
      if _, ok := tag_map[v]; !ok {
        tag_list = append(tag_list, v)
        tag_map[v] = true
      }
    }
    tags = tag_list
  } else {
    tags = []string{"untagged"}
    only_untagged = true
  }

  if len(only_tags) > 0 {
    should_run = false
    if StringPos("always", tags) != -1 || StringPos("all", only_tags) != -1 {
      should_run = true
    } else if !Disjoint(tags, only_tags) {
      should_run = true
    } else if StringPos("tagged", only_tags) != -1 && !only_untagged {
      should_run = true
    }
  }
  if should_run && len(skip_tags) > 0 {
    // Check for tags that we need to skip
    if StringPos("all", skip_tags) != -1 {
      if StringPos("always", tags) == -1 || StringPos("always", skip_tags) != -1 {
        should_run = false
      }
    } else if !Disjoint(tags, skip_tags) {
      should_run = false
    } else if StringPos("tagged", skip_tags) != -1 && !only_untagged {
      should_run = false
    }
  }

  return should_run
}
