#!/usr/bin/python
# -*- coding: utf-8 -*-

# Copyright: (c) 2017, Bruno Calogero <brunocalogero@hotmail.com>
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function
__metaclass__ = type

ANSIBLE_METADATA = {'metadata_version': '1.1',
                    'status': ['preview'],
                    'supported_by': 'community'}

DOCUMENTATION = r'''
---
module: aci_access_port_to_interface_policy_leaf_profile
short_description: Manage Fabric interface policy leaf profile interface selectors on Cisco ACI fabrics (infra:HPortS, infra:RsAccBaseGrp, infra:PortBlk)
description:
- Manage Fabric interface policy leaf profile interface selectors on Cisco ACI fabrics.
- More information from the internal APIC class I(infra:HPortS, infra:RsAccBaseGrp, infra:PortBlk) at
  U(https://developer.cisco.com/media/mim-ref).
author:
- Bruno Calogero (@brunocalogero)
version_added: '2.5'
options:
  leaf_interface_profile:
    description:
    - The name of the Fabric access policy leaf interface profile.
    required: yes
    aliases: [ leaf_interface_profile_name ]
  access_port_selector:
    description:
    -  The name of the Fabric access policy leaf interface profile access port selector.
    required: yes
    aliases: [ name, access_port_selector_name ]
  description:
    description:
    - The description to assign to the C(access_port_selector)
    required: no
  leaf_port_blk:
    description:
    - The name of the Fabric access policy leaf interface profile access port block.
    required: yes
    aliases: [ leaf_port_blk_name ]
  leaf_port_blk_description:
    description:
    - The description to assign to the C(leaf_port_blk)
    required: no
  from:
    description:
    - The beggining (from range) of the port range block for the leaf access port block.
    required: yes
    aliases: [ fromPort, from_port_range ]
  to:
    description:
    - The end (to range) of the port range block for the leaf access port block.
    required: yes
    aliases: [ toPort, to_port_range ]
  policy_group:
    description:
    - The name of the fabric access policy group to be associated with the leaf interface profile interface selector.
    required: no
    aliases: [ policy_group_name ]
  state:
    description:
    - Use C(present) or C(absent) for adding or removing.
    - Use C(query) for listing an object or multiple objects.
    choices: [ absent, present, query ]
    default: present
'''

EXAMPLES = r'''
- name: Associate an Interface Access Port Selector to an Interface Policy Leaf Profile with a Policy Group
  aci_access_port_to_interface_policy_leaf_profile:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    leaf_interface_profile: leafintprfname
    access_port_selector: accessportselectorname
    leaf_port_blk: leafportblkname
    from: 13
    to: 16
    policy_group: policygroupname
    state: present

- name: Associate an interface access port selector to an Interface Policy Leaf Profile (w/o policy group) (check if this works)
  aci_access_port_to_interface_policy_leaf_profile:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    leaf_interface_profile: leafintprfname
    access_port_selector: accessportselectorname
    leaf_port_blk: leafportblkname
    from: 13
    to: 16
    state: present

- name: Remove an interface access port selector associated with an Interface Policy Leaf Profile
  aci_access_port_to_interface_policy_leaf_profile:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    leaf_interface_profile: leafintprfname
    access_port_selector: accessportselectorname
    state: absent

- name: Query Specific access_port_selector under given leaf_interface_profile
  aci_access_port_to_interface_policy_leaf_profile:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    leaf_interface_profile: leafintprfname
    access_port_selector: accessportselectorname
    state: query
'''

RETURN = r'''
#
'''

from ansible.module_utils.network.aci.aci import ACIModule, aci_argument_spec
from ansible.module_utils.basic import AnsibleModule


def main():
    argument_spec = aci_argument_spec
    argument_spec.update({
        'leaf_interface_profile': dict(type='str', aliases=['leaf_interface_profile_name']),
        'access_port_selector': dict(type='str', aliases=['name', 'access_port_selector_name']),
        'description': dict(typ='str'),
        'leaf_port_blk': dict(type='str', aliases=['leaf_port_blk_name']),
        'leaf_port_blk_description': dict(type='str'),
        'from': dict(type='str', aliases=['fromPort', 'from_port_range']),
        'to': dict(type='str', aliases=['toPort', 'to_port_range']),
        'policy_group': dict(type='str', aliases=['policy_group_name']),
        'state': dict(type='str', default='present', choices=['absent', 'present', 'query']),
    })

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
        required_if=[
            ['state', 'absent', ['leaf_interface_profile', 'access_port_selector']],
            ['state', 'present', ['leaf_interface_profile', 'access_port_selector']],
        ],
    )

    leaf_interface_profile = module.params['leaf_interface_profile']
    access_port_selector = module.params['access_port_selector']
    description = module.params['description']
    leaf_port_blk = module.params['leaf_port_blk']
    leaf_port_blk_description = module.params['leaf_port_blk_description']
    from_ = module.params['from']
    to_ = module.params['to']
    policy_group = module.params['policy_group']
    state = module.params['state']

    aci = ACIModule(module)
    aci.construct_url(
        root_class=dict(
            aci_class='infraAccPortP',
            aci_rn='infra/accportprof-{0}'.format(leaf_interface_profile),
            filter_target='eq(infraAccPortP.name, "{0}")'.format(leaf_interface_profile),
            module_object=leaf_interface_profile
        ),
        subclass_1=dict(
            aci_class='infraHPortS',
            # NOTE: normal rn: hports-{name}-typ-{type}, hence here hardcoded to range for purposes of module
            aci_rn='hports-{0}-typ-range'.format(access_port_selector),
            filter_target='eq(infraHPortS.name, "{0}")'.format(access_port_selector),
            module_object=access_port_selector,
        ),
        child_classes=['infraPortBlk', 'infraRsAccBaseGrp']
    )
    aci.get_existing()

    if state == 'present':
        # Filter out module parameters with null values
        aci.payload(
            aci_class='infraHPortS',
            class_config=dict(
                descr=description,
                name=access_port_selector,
            ),
            child_configs=[
                dict(
                    infraPortBlk=dict(
                        attributes=dict(
                            descr=leaf_port_blk_description,
                            name=leaf_port_blk,
                            fromPort=from_,
                            toPort=to_,
                        )
                    )
                ),
                dict(
                    infraRsAccBaseGrp=dict(
                        attributes=dict(
                            tDn='uni/infra/funcprof/accportgrp-{0}'.format(policy_group),
                        )
                    )
                ),
            ],
        )

        # Generate config diff which will be used as POST request body
        aci.get_diff(aci_class='infraHPortS')

        # Submit changes if module not in check_mode and the proposed is different than existing
        aci.post_config()

    elif state == 'absent':
        aci.delete_config()

    module.exit_json(**aci.result)


if __name__ == "__main__":
    main()
