#!/usr/bin/python
# -*- coding: utf-8 -*-

# (c) 2017, Ansible by Red Hat, inc
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function
__metaclass__ = type


ANSIBLE_METADATA = {'metadata_version': '1.1',
                    'status': ['preview'],
                    'supported_by': 'network'}


DOCUMENTATION = """
---
module: nxos_l3_interface
version_added: "2.5"
author: "Trishna Guha (@trishnaguha)"
short_description: Manage L3 interfaces on Cisco NXOS network devices
description:
  - This module provides declarative management of L3 interfaces
    on Cisco NXOS network devices.
notes:
  - Tested against NXOSv 7.0(3)I5(1).
options:
  name:
    description:
      - Name of the L3 interface.
  ipv4:
    description:
      - IPv4 of the L3 interface.
  ipv6:
    description:
      - IPv6 of the L3 interface.
  aggregate:
    description: List of L3 interfaces definitions.
  state:
    description:
      - State of the L3 interface configuration.
    default: present
    choices: ['present', 'absent']
"""

EXAMPLES = """
- name: Set interface IPv4 address
  nxos_l3_interface:
    name: Ethernet2/3
    ipv4: 192.168.0.1/24

- name: Remove interface IPv4 address
  nxos_l3_interface:
    name: Ethernet2/3
    state: absent

- name: Set IP addresses on aggregate
  nxos_l3_interface:
    aggregate:
      - { name: Ethernet2/1, ipv4: 192.168.2.10/24 }
      - { name: Ethernet2/5, ipv4: 192.168.3.10/24, ipv6: "fd5d:12c9:2201:1::1/64" }

- name: Remove IP addresses on aggregate
  nxos_l3_interface:
    aggregate:
      - { name: Ethernet2/1, ipv4: 192.168.2.10/24 }
      - { name: Ethernet2/5, ipv4: 192.168.3.10/24, ipv6: "fd5d:12c9:2201:1::1/64" }
    state: absent
"""

RETURN = """
commands:
  description: The list of configuration mode commands to send to the device
  returned: always, except for the platforms that use Netconf transport to manage the device.
  type: list
  sample:
    - interface ethernet2/3
    - no switchport
    - ip address 192.168.22.1/24
    - ipv6 address "fd5d:12c9:2201:1::1/64"
    - no ip address 192.168.22.1/24
"""

import re

from copy import deepcopy

from ansible.module_utils.basic import AnsibleModule
from ansible.module_utils.network.common.config import CustomNetworkConfig
from ansible.module_utils.network.common.utils import remove_default_spec
from ansible.module_utils.network.nxos.nxos import get_config, load_config
from ansible.module_utils.network.nxos.nxos import nxos_argument_spec


def search_obj_in_list(name, lst):
    for o in lst:
        if o['name'] == name:
            return o


def map_obj_to_commands(updates, module):
    commands = list()
    want, have = updates

    for w in want:
        name = w['name']
        ipv4 = w['ipv4']
        ipv6 = w['ipv6']
        state = w['state']
        del w['state']

        obj_in_have = search_obj_in_list(name, have)

        if state == 'absent' and obj_in_have:
            command = []
            if obj_in_have['name'] == name:
                if ipv4 and obj_in_have['ipv4']:
                    command.append('no ip address {0}'.format(ipv4))
                if ipv6 and obj_in_have['ipv6']:
                    command.append('no ipv6 address {0}'.format(ipv6))
                if command:
                    command.append('exit')
                    command.insert(0, 'interface {0}'.format(name))
                    command.insert(1, 'no switchport')
            commands.extend(command)

        elif state == 'present' and obj_in_have:
            command = []
            if obj_in_have['name'] == name:
                if ipv4 and ipv4 != obj_in_have['ipv4']:
                    command.append('ip address {0}'.format(ipv4))
                if ipv6 and ipv6 != obj_in_have['ipv6']:
                    command.append('ipv6 address {0}'.format(ipv6))
                if command:
                    command.append('exit')
                    command.insert(0, 'interface {0}'.format(name))
                    command.insert(1, 'no switchport')
                elif not ipv4 and not ipv6:
                    command.append('interface {0}'.format(name))
                    command.append('no switchport')
            commands.extend(command)

    return commands


def map_params_to_obj(module):
    obj = []

    aggregate = module.params.get('aggregate')
    if aggregate:
        for item in aggregate:
            for key in item:
                if item.get(key) is None:
                    item[key] = module.params[key]

            obj.append(item.copy())
    else:
        obj.append({
            'name': module.params['name'],
            'ipv4': module.params['ipv4'],
            'ipv6': module.params['ipv6'],
            'state': module.params['state']
        })

    return obj


def map_config_to_obj(want, module):
    objs = list()
    netcfg = CustomNetworkConfig(indent=2, contents=get_config(module))

    for w in want:
        parents = ['interface {0}'.format(w['name'])]
        config = netcfg.get_section(parents)
        obj = dict(name=None, ipv4=None, ipv6=None)

        if config:
            match_name = re.findall(r'interface (\S+)', config, re.M)
            if match_name:
                obj['name'] = match_name[0]

            match_ipv4 = re.findall(r'ip address (\S+)', config, re.M)
            if match_ipv4:
                obj['ipv4'] = match_ipv4[0]

            match_ipv6 = re.findall(r'ipv6 address (\S+)', config, re.M)
            if match_ipv6:
                obj['ipv6'] = match_ipv6[0]

        objs.append(obj)
    return objs


def main():
    """ main entry point for module execution
    """
    element_spec = dict(
        name=dict(),
        ipv4=dict(),
        ipv6=dict(),
        state=dict(default='present', choices=['present', 'absent'])
    )

    aggregate_spec = deepcopy(element_spec)

    # remove default in aggregate spec, to handle common arguments
    remove_default_spec(aggregate_spec)

    argument_spec = dict(
        aggregate=dict(type='list', elements='dict', options=aggregate_spec),
    )

    argument_spec.update(element_spec)
    argument_spec.update(nxos_argument_spec)

    required_one_of = [['name', 'aggregate']]
    mutually_exclusive = [['name', 'aggregate']]
    module = AnsibleModule(argument_spec=argument_spec,
                           required_one_of=required_one_of,
                           mutually_exclusive=mutually_exclusive,
                           supports_check_mode=True)

    warnings = list()
    result = {'changed': False}
    if warnings:
        result['warnings'] = warnings

    want = map_params_to_obj(module)
    have = map_config_to_obj(want, module)

    commands = map_obj_to_commands((want, have), module)
    result['commands'] = commands

    if commands:
        if not module.check_mode:
            load_config(module, commands)
        result['changed'] = True

    module.exit_json(**result)

if __name__ == '__main__':
    main()
