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
module: aci_switch_leaf_selector
short_description: Add a leaf Selector with Node Block Range and Policy Group to a Switch Policy Leaf Profile on Cisco ACI fabrics
description:
- Add a leaf Selector with Node Block range and Policy Group to a Switch Policy Leaf Profile on Cisco ACI fabrics.
- More information from the internal APIC class I(infra:LeafS), I(infra:NodeBlk), I(infra:RsAccNodePGrp) at
  U(https://developer.cisco.com/docs/apic-mim-ref/).
author:
- Bruno Calogero (@brunocalogero)
version_added: '2.5'
notes:
- This module is to be used with M(aci_switch_policy_leaf_profile)
  One first creates a leaf profile (infra:NodeP) and then creates an associated selector (infra:LeafS),
options:
  description:
    description:
    - The description to assign to the C(leaf)
  leaf_profile:
    description:
    - Name of the Leaf Profile to which we add a Selector.
    aliases: [ leaf_profile_name ]
  leaf:
    description:
    - Name of Leaf Selector.
    aliases: [ name, leaf_name, leaf_profile_leaf_name, leaf_selector_name ]
  leaf_node_blk:
    description:
    - Name of Node Block range to be added to Leaf Selector of given Leaf Profile
    aliases: [ leaf_node_blk_name, node_blk_name ]
  leaf_node_blk_description:
    description:
    - The description to assign to the C(leaf_node_blk)
  from:
    description:
    - Start of Node Block Range
    aliases: [ node_blk_range_from, from_range, range_from ]
  to:
    description:
    - Start of Node Block Range
    aliases: [ node_blk_range_to, to_range, range_to ]
  policy_group:
    description:
    - Name of the Policy Group to be added to Leaf Selector of given Leaf Profile
    aliases: [ name, policy_group_name ]
  state:
    description:
    - Use C(present) or C(absent) for adding or removing.
    - Use C(query) for listing an object or multiple objects.
    choices: [ absent, present, query ]
    default: present
'''

EXAMPLES = r'''
- name: adding a switch policy leaf profile selector associated Node Block range (w/ policy group)
  aci_switch_leaf_selector:
    hostname: apic
    username: someusername
    password: somepassword
    leaf_profile: sw_name
    leaf: leaf_selector_name
    leaf_node_blk: node_blk_name
    from: 1011
    to: 1011
    policy_group: somepolicygroupname
    state: present

- name: adding a switch policy leaf profile selector associated Node Block range (w/o policy group)
  aci_switch_leaf_selector:
    hostname: apic
    username: someusername
    password: somepassword
    leaf_profile: sw_name
    leaf: leaf_selector_name
    leaf_node_blk: node_blk_name
    from: 1011
    to: 1011
    state: present

- name: Removing a switch policy leaf profile selector
  aci_switch_leaf_selector:
    hostname: apic
    username: someusername
    password: somepassword
    leaf_profile: sw_name
    leaf: leaf_selector_name
    state: absent

- name: Querying a switch policy leaf profile selector
  aci_switch_leaf_selector:
    hostname: apic
    username: someusername
    password: somepassword
    leaf_profile: sw_name
    leaf: leaf_selector_name
    state: query
'''

RETURN = ''' # '''

from ansible.module_utils.network.aci.aci import ACIModule, aci_argument_spec
from ansible.module_utils.basic import AnsibleModule


def main():
    argument_spec = aci_argument_spec
    argument_spec.update({
        'description': dict(type='str'),
        'leaf_profile': dict(type='str', aliases=['leaf_profile_name']),
        'leaf': dict(type='str', aliases=['name', 'leaf_name', 'leaf_profile_leaf_name', 'leaf_selector_name']),
        'leaf_node_blk': dict(type='str', aliases=['leaf_node_blk_name', 'node_blk_name']),
        'leaf_node_blk_description': dict(type='str'),
        'from': dict(type='int', aliases=['node_blk_range_from', 'from_range', 'range_from']),
        'to': dict(type='int', aliases=['node_blk_range_to', 'to_range', 'range_to']),
        'policy_group': dict(type='str', aliases=['policy_group_name']),
        'state': dict(type='str', default='present', choices=['absent', 'present', 'query']),
    })

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
        required_if=[
            ['state', 'absent', ['leaf_profile', 'leaf']],
            ['state', 'present', ['leaf_profile', 'leaf', 'leaf_node_blk', 'from', 'to']]
        ]
    )

    description = module.params['description']
    leaf_profile = module.params['leaf_profile']
    leaf = module.params['leaf']
    leaf_node_blk = module.params['leaf_node_blk']
    leaf_node_blk_description = module.params['leaf_node_blk_description']
    from_ = module.params['from']
    to_ = module.params['to']
    policy_group = module.params['policy_group']
    state = module.params['state']

    aci = ACIModule(module)
    aci.construct_url(
        root_class=dict(
            aci_class='infraNodeP',
            aci_rn='infra/nprof-{0}'.format(leaf_profile),
            filter_target='eq(infraNodeP.name, "{0}")'.format(leaf_profile),
            module_object=leaf_profile
        ),
        subclass_1=dict(
            aci_class='infraLeafS',
            # NOTE: normal rn: leaves-{name}-typ-{type}, hence here hardcoded to range for purposes of module
            aci_rn='leaves-{0}-typ-range'.format(leaf),
            filter_target='eq(infraLeafS.name, "{0}")'.format(leaf),
            module_object=leaf,
        ),
        # NOTE: infraNodeBlk is not made into a subclass because there is a 1-1 mapping between node block and leaf selector name
        child_classes=['infraNodeBlk', 'infraRsAccNodePGrp']

    )

    aci.get_existing()

    if state == 'present':
        # Filter out module params with null values
        aci.payload(
            aci_class='infraLeafS',
            class_config=dict(
                descr=description,
                name=leaf,
            ),
            child_configs=[
                dict(
                    infraNodeBlk=dict(
                        attributes=dict(
                            descr=leaf_node_blk_description,
                            name=leaf_node_blk,
                            from_=from_,
                            to_=to_,
                        )
                    )
                ),
                dict(
                    infraRsAccNodePGrp=dict(
                        attributes=dict(
                            tDn='uni/infra/funcprof/accnodepgrp-{0}'.format(policy_group),
                        )
                    )
                ),
            ],
        )

        # Generate config diff which will be used as POST request body
        aci.get_diff(aci_class='infraLeafS')

        # Submit changes if module not in check_mode and the proposed is different than existing
        aci.post_config()

    elif state == 'absent':
        aci.delete_config()

    module.exit_json(**aci.result)


if __name__ == "__main__":
    main()
