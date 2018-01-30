package playbook

import ()

var TASK_ATTRIBUTE_OVERRIDES = []string{
  "become",
  "become_user",
  "become_pass",
  "become_method",
  "become_flags",
  "connection",
  "docker_extra_args",  // TODO: remove
  "delegate_to",
  "no_log",
  "remote_user",
}

var RESET_VARS = []string{
  "ansible_connection",
  "ansible_user",
  "ansible_host",
  "ansible_port",
  // TODO: ???
  "ansible_docker_extra_args",
  "ansible_ssh_host",
  "ansible_ssh_pass",
  "ansible_ssh_port",
  "ansible_ssh_user",
  "ansible_ssh_private_key_file",
  "ansible_ssh_pipelining",
  "ansible_ssh_executable",
}

var OPTION_FLAGS = []string{
  "connection",
  "remote_user",
  "private_key_file",
  "verbosity",
  "force_handlers",
  "step",
  "start_at_task",
  "diff",
  "ssh_common_args",
  "docker_extra_args",
  "sftp_extra_args",
  "scp_extra_args",
  "ssh_extra_args",
}

var play_context_fields = map[string]FieldAttribute{
  "skip_tags": FieldAttribute{T:"list", Default: nil, ListOf: "strings"},
  "only_tags": FieldAttribute{T:"list", Default: nil, ListOf: "strings"},
}

type PlayContext struct {
  Base

  Attr_module_compression interface{}
  Attr_shell interface{}
  Attr_executable interface{}
  // connection fields, some are inherited from Base:
  // (connection, port, remote_user, environment, no_log)
  Attr_remote_addr interface{}
  Attr_password interface{}
  Attr_timeout interface{}
  Attr_connection_user interface{}
  Attr_private_key_file interface{}
  Attr_pipelining interface{}
  // networking modules
  Attr_network_os interface{}
  // docker FIXME: remove these
  Attr_docker_extra_args interface{}
  // ssh # FIXME: remove these
  Attr_ssh_executable interface{}
  Attr_ssh_args interface{}
  Attr_ssh_common_args interface{}
  Attr_sftp_extra_args interface{}
  Attr_scp_extra_args interface{}
  Attr_ssh_extra_args interface{}
  Attr_ssh_transfer_method interface{}
  // privilege escalation fields
  Attr_become interface{}
  Attr_become_method interface{}
  Attr_become_user interface{}
  Attr_become_pass interface{}
  Attr_become_exe interface{}
  Attr_become_flags interface{}
  Attr_prompt interface{}
  // DEPRECATED: backwards compatibility fields for sudo/su
  Attr_sudo_exe interface{}
  Attr_sudo_flags interface{}
  Attr_sudo_pass interface{}
  Attr_su_exe interface{}
  Attr_su_flags interface{}
  Attr_su_pass interface{}
  // general flags
  Attr_only_tags interface{}
  Attr_skip_tags interface{}
  Attr_verbosity interface{}
  Attr_force_handlers interface{}
  Attr_start_at_task interface{}
  Attr_step interface{}
  // Fact gathering settings
  Attr_gather_subset interface{}
  Attr_gather_timeout interface{}
  Attr_fact_path interface{}

  //_connection_lockfd = FieldAttribute(isa='int')
}

func (pc *PlayContext) GetAllObjectFieldAttributes() map[string]FieldAttribute {
  var all_fields = make(map[string]FieldAttribute)
  var items = []map[string]FieldAttribute{base_fields, play_context_fields}
  for i := 0; i < len(items); i++ {
    for k, v := range items[i] {
      all_fields[k] = v
    }
  }
  return all_fields
}

func (pc *PlayContext) Load(play *Play, options interface{}) {
  dummy := make(map[interface{}]interface{})

  pc.Base.Load(dummy)
  pc.Base.GetInheritedValue = pc.GetInheritedValue
  pc.Base.GetAllObjectFieldAttributes = pc.GetAllObjectFieldAttributes

  // we use LoadValidFields() here to initialize everything to defaults
  LoadValidFields(pc, play_context_fields, dummy)
  if options != nil {
    pc.SetOptions(options)
  }
  if play != nil {
    pc.SetPlay(play)
  }
}

func (pc *PlayContext) SetOptions(options interface{}) {
}

func (pc *PlayContext) SetPlay(play *Play) {
}

// local getters
func (pc *PlayContext) Skip_tags() []string {
  if res, ok := pc.Attr_skip_tags.([]string); ok {
    return res
  } else {
    res, _ := play_context_fields["skip_tags"].Default.([]string)
    return res
  }
}
func (pc *PlayContext) Only_tags() []string {
  if res, ok := pc.Attr_only_tags.([]string); ok {
    return res
  } else {
    res, _ := play_context_fields["only_tags"].Default.([]string)
    return res
  }
}

// the new method
func NewPlayContext(play *Play, options interface{}, passwords []string) *PlayContext {
  pc := new(PlayContext)
  pc.Load(play, options)
  return pc
}
