#!/usr/bin/python
# -*- coding: utf-8 -*-

# Copyright: (c) 2014, 2015 YAEGASHI Takeshi <yaegashi@debian.org>
# Copyright: (c) 2017, Ansible Project
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function
__metaclass__ = type

ANSIBLE_METADATA = {'metadata_version': '1.1',
                    'status': ['preview'],
                    'supported_by': 'core'}


DOCUMENTATION = """
---
module: blockinfile
author:
    - YAEGASHI Takeshi (@yaegashi)
extends_documentation_fragment:
    - files
    - validate
short_description: Insert/update/remove a text block surrounded by marker lines
version_added: '2.0'
description:
  - This module will insert/update/remove a block of multi-line text
    surrounded by customizable marker lines.
options:
  path:
    description:
      - The file to modify.
      - Before 2.3 this option was only usable as I(dest), I(destfile) and I(name).
    aliases: [ dest, destfile, name ]
    required: true
  state:
    description:
      - Whether the block should be there or not.
    choices: [ absent, present ]
    default: present
  marker:
    description:
      - The marker line template.
        "{mark}" will be replaced with the values in marker_begin
        (default="BEGIN") and marker_end (default="END").
    default: '# {mark} ANSIBLE MANAGED BLOCK'
  block:
    description:
      - The text to insert inside the marker lines.
        If it's missing or an empty string,
        the block will be removed as if C(state) were specified to C(absent).
    aliases: [ content ]
    default: ''
  insertafter:
    description:
      - If specified, the block will be inserted after the last match of
        specified regular expression. A special value is available; C(EOF) for
        inserting the block at the end of the file.  If specified regular
        expression has no matches, C(EOF) will be used instead.
    default: EOF
    choices: [ EOF, '*regex*' ]
  insertbefore:
    description:
      - If specified, the block will be inserted before the last match of
        specified regular expression. A special value is available; C(BOF) for
        inserting the block at the beginning of the file.  If specified regular
        expression has no matches, the block will be inserted at the end of the
        file.
    choices: [ BOF, '*regex*' ]
  create:
    description:
      - Create a new file if it doesn't exist.
    type: bool
    default: 'no'
  backup:
    description:
      - Create a backup file including the timestamp information so you can
        get the original file back if you somehow clobbered it incorrectly.
    type: bool
    default: 'no'
  marker_begin:
    description:
      - This will be inserted at {mark} in the opening ansible block marker.
    default: 'BEGIN'
    version_added: "2.5"
  marker_end:
    required: false
    description:
      - This will be inserted at {mark} in the closing ansible block marker.
    default: 'END'
    version_added: "2.5"

notes:
  - This module supports check mode.
  - When using 'with_*' loops be aware that if you do not set a unique mark the block will be overwritten on each iteration.
  - As of Ansible 2.3, the I(dest) option has been changed to I(path) as default, but I(dest) still works as well.
  - Option I(follow) has been removed in version 2.5, because this module modifies the contents of the file so I(follow=no) doesn't make sense.
"""

EXAMPLES = r"""
# Before 2.3, option 'dest' or 'name' was used instead of 'path'
- name: insert/update "Match User" configuration block in /etc/ssh/sshd_config
  blockinfile:
    path: /etc/ssh/sshd_config
    block: |
      Match User ansible-agent
      PasswordAuthentication no

- name: insert/update eth0 configuration stanza in /etc/network/interfaces
        (it might be better to copy files into /etc/network/interfaces.d/)
  blockinfile:
    path: /etc/network/interfaces
    block: |
      iface eth0 inet static
          address 192.0.2.23
          netmask 255.255.255.0

- name: insert/update configuration using a local file and validate it
  blockinfile:
    block: "{{ lookup('file', './local/ssh_config') }}"
    dest: "/etc/ssh/ssh_config"
    backup: yes
    validate: "/usr/sbin/sshd -T -f %s"

- name: insert/update HTML surrounded by custom markers after <body> line
  blockinfile:
    path: /var/www/html/index.html
    marker: "<!-- {mark} ANSIBLE MANAGED BLOCK -->"
    insertafter: "<body>"
    content: |
      <h1>Welcome to {{ ansible_hostname }}</h1>
      <p>Last updated on {{ ansible_date_time.iso8601 }}</p>

- name: remove HTML as well as surrounding markers
  blockinfile:
    path: /var/www/html/index.html
    marker: "<!-- {mark} ANSIBLE MANAGED BLOCK -->"
    content: ""

- name: Add mappings to /etc/hosts
  blockinfile:
    path: /etc/hosts
    block: |
      {{ item.ip }} {{ item.name }}
    marker: "# {mark} ANSIBLE MANAGED BLOCK {{ item.name }}"
  with_items:
    - { name: host1, ip: 10.10.1.10 }
    - { name: host2, ip: 10.10.1.11 }
    - { name: host3, ip: 10.10.1.12 }
"""

import re
import os
import tempfile
from ansible.module_utils.six import b
from ansible.module_utils.basic import AnsibleModule
from ansible.module_utils._text import to_bytes


def write_changes(module, contents, path):

    tmpfd, tmpfile = tempfile.mkstemp()
    f = os.fdopen(tmpfd, 'wb')
    f.write(contents)
    f.close()

    validate = module.params.get('validate', None)
    valid = not validate
    if validate:
        if "%s" not in validate:
            module.fail_json(msg="validate must contain %%s: %s" % (validate))
        (rc, out, err) = module.run_command(validate % tmpfile)
        valid = rc == 0
        if rc != 0:
            module.fail_json(msg='failed to validate: '
                                 'rc:%s error:%s' % (rc, err))
    if valid:
        module.atomic_move(tmpfile, path, unsafe_writes=module.params['unsafe_writes'])


def check_file_attrs(module, changed, message, diff):

    file_args = module.load_file_common_arguments(module.params)
    if module.set_file_attributes_if_different(file_args, False, diff=diff):

        if changed:
            message += " and "
        changed = True
        message += "ownership, perms or SE linux context changed"

    return message, changed


def main():
    module = AnsibleModule(
        argument_spec=dict(
            path=dict(type='path', required=True, aliases=['dest', 'destfile', 'name']),
            state=dict(type='str', default='present', choices=['absent', 'present']),
            marker=dict(type='str', default='# {mark} ANSIBLE MANAGED BLOCK'),
            block=dict(type='str', default='', aliases=['content']),
            insertafter=dict(type='str'),
            insertbefore=dict(type='str'),
            create=dict(type='bool', default=False),
            backup=dict(type='bool', default=False),
            validate=dict(type='str'),
            marker_begin=dict(type='str', default='BEGIN'),
            marker_end=dict(type='str', default='END'),
        ),
        mutually_exclusive=[['insertbefore', 'insertafter']],
        add_file_common_args=True,
        supports_check_mode=True
    )

    params = module.params
    path = params['path']

    if os.path.isdir(path):
        module.fail_json(rc=256,
                         msg='Path %s is a directory !' % path)

    path_exists = os.path.exists(path)
    if not path_exists:
        if not module.boolean(params['create']):
            module.fail_json(rc=257,
                             msg='Path %s does not exist !' % path)
        original = None
        lines = []
    else:
        f = open(path, 'rb')
        original = f.read()
        f.close()
        lines = original.splitlines()

    diff = {'before': '',
            'after': '',
            'before_header': '%s (content)' % path,
            'after_header': '%s (content)' % path}

    if module._diff and original:
        diff['before'] = original

    insertbefore = params['insertbefore']
    insertafter = params['insertafter']
    block = to_bytes(params['block'])
    marker = to_bytes(params['marker'])
    present = params['state'] == 'present'

    if not present and not path_exists:
        module.exit_json(changed=False, msg="File %s not present" % path)

    if insertbefore is None and insertafter is None:
        insertafter = 'EOF'

    if insertafter not in (None, 'EOF'):
        insertre = re.compile(to_bytes(insertafter, errors='surrogate_or_strict'))
    elif insertbefore not in (None, 'BOF'):
        insertre = re.compile(to_bytes(insertbefore, errors='surrogate_or_strict'))
    else:
        insertre = None

    marker0 = re.sub(b(r'{mark}'), b(params['marker_begin']), marker)
    marker1 = re.sub(b(r'{mark}'), b(params['marker_end']), marker)
    if present and block:
        # Escape seqeuences like '\n' need to be handled in Ansible 1.x
        if module.ansible_version.startswith('1.'):
            block = re.sub('', block, '')
        blocklines = [marker0] + block.splitlines() + [marker1]
    else:
        blocklines = []

    n0 = n1 = None
    for i, line in enumerate(lines):
        if line == marker0:
            n0 = i
        if line == marker1:
            n1 = i

    if None in (n0, n1):
        n0 = None
        if insertre is not None:
            for i, line in enumerate(lines):
                if insertre.search(line):
                    n0 = i
            if n0 is None:
                n0 = len(lines)
            elif insertafter is not None:
                n0 += 1
        elif insertbefore is not None:
            n0 = 0  # insertbefore=BOF
        else:
            n0 = len(lines)  # insertafter=EOF
    elif n0 < n1:
        lines[n0:n1 + 1] = []
    else:
        lines[n1:n0 + 1] = []
        n0 = n1

    lines[n0:n0] = blocklines

    if lines:
        result = b('\n').join(lines)
        if original is None or original.endswith(b('\n')):
            result += b('\n')
    else:
        result = ''

    if module._diff:
        diff['after'] = result

    if original == result:
        msg = ''
        changed = False
    elif original is None:
        msg = 'File created'
        changed = True
    elif not blocklines:
        msg = 'Block removed'
        changed = True
    else:
        msg = 'Block inserted'
        changed = True

    if changed and not module.check_mode:
        if module.boolean(params['backup']) and path_exists:
            module.backup_local(path)
        # We should always follow symlinks so that we change the real file
        real_path = os.path.realpath(params['path'])
        write_changes(module, result, real_path)

    if module.check_mode and not path_exists:
        module.exit_json(changed=changed, msg=msg, diff=diff)

    attr_diff = {}
    msg, changed = check_file_attrs(module, changed, msg, attr_diff)

    attr_diff['before_header'] = '%s (file attributes)' % path
    attr_diff['after_header'] = '%s (file attributes)' % path

    difflist = [diff, attr_diff]
    module.exit_json(changed=changed, msg=msg, diff=difflist)


if __name__ == '__main__':
    main()
