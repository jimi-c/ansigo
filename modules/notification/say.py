#!/usr/bin/python
# -*- coding: utf-8 -*-

# (c) 2013, Michael DeHaan <michael@ansible.com>
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function
__metaclass__ = type


ANSIBLE_METADATA = {'metadata_version': '1.1',
                    'status': ['stableinterface'],
                    'supported_by': 'community'}


DOCUMENTATION = '''
---
module: say
version_added: "1.2"
short_description: Makes a computer to speak.
description:
   - makes a computer speak! Amuse your friends, annoy your coworkers!
notes:
   - In 2.5, this module has been renamed from M(osx_say) into M(say).
   - If you like this module, you may also be interested in the osx_say callback plugin.
options:
  msg:
    description:
      What to say
    required: true
  voice:
    description:
      What voice to use
    required: false
requirements: [ say or espeak ]
author:
    - "Ansible Core Team"
    - "Michael DeHaan (@mpdehaan)"
'''

EXAMPLES = '''
- say:
    msg: '{{ inventory_hostname }} is all done'
    voice: Zarvox
  delegate_to: localhost
'''
import os

from ansible.module_utils.basic import AnsibleModule, get_platform


def say(module, executable, msg, voice):
    cmd = [executable, msg]
    if voice:
        cmd.extend(('-v', voice))
    module.run_command(cmd, check_rc=True)


def main():

    module = AnsibleModule(
        argument_spec=dict(
            msg=dict(required=True),
            voice=dict(required=False),
        ),
        supports_check_mode=True
    )

    msg = module.params['msg']
    voice = module.params['voice']

    executable = module.get_bin_path('say')
    if not executable:
        executable = module.get_bin_path('espeak')
    elif get_platform() != 'Darwin':
        # 'say' binary available, it might be GNUstep tool which doesn't support 'voice' parameter
        voice = None
        module.warn("'say' executable found but system is '%s': ignoring voice parameter" % get_platform())

    if not executable:
        module.fail_json(msg="Unable to find either 'say' or 'espeak' executable")

    if module.check_mode:
        module.exit_json(msg=msg, changed=False)

    say(module, executable, msg, voice)

    module.exit_json(msg=msg, changed=True)


if __name__ == '__main__':
    main()
