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
module: aci_interface_selector_to_switch_policy_leaf_profile
short_description: Associates an Interface Selector Profile to a Switch Policy Leaf Profile (infra:RsAccPortP)
description:
- Associates an Interface Profile (Selector) to a Switch Policy Leaf Profile on Cisco ACI fabrics.
- More information from the internal APIC class I(infra:RsAccPortP) at
  U(https://developer.cisco.com/docs/apic-mim-ref/).
author:
- Bruno Calogero (@brunocalogero)
version_added: '2.5'
notes:
- This module can be used with M(aci_switch_policy_leaf_profile).
  One first creates a leaf profile (infra:NodeP),
  Finally, associates an interface profile using the provided interface selector profile (infra:RsAccPortP)
options:
  leaf_profile:
    description:
    - Name of the Leaf Profile to which we add a Selector.
    aliases: [ leaf_profile_name ]
  interface_selector:
    description:
    - Name of Interface Profile Selector to be added and associated with the Leaf Profile.
    aliases: [ name, interface_selector_name, interface_profile_name ]
  state:
    description:
    - Use C(present) or C(absent) for adding or removing.
    - Use C(query) for listing an object or multiple objects.
    choices: [ absent, present, query ]
    default: present
'''

EXAMPLES = r'''
- name: Associating an interface selector profile to a switch policy leaf profile
  aci_interface_selector_to_switch_policy_leaf_profile:
    hostname: apic
    username: someusername
    password: somepassword
    leaf_profile: sw_name
    interface_selector: interface_profile_name
    state: present

- name: Remove an interface selector profile associated with a switch policy leaf profile
  aci_interface_selector_to_switch_policy_leaf_profile:
    hostname: apic
    username: someusername
    password: somepassword
    leaf_profile: sw_name
    interface_selector: interface_profile_name
    state: absent

- name: Query an interface selector profile associated with a switch policy leaf profile
  aci_interface_selector_to_switch_policy_leaf_profile:
    hostname: apic
    username: someusername
    password: somepassword
    leaf_profile: sw_name
    interface_selector: interface_profile_name
    state: query
'''

RETURN = ''' # '''

from ansible.module_utils.network.aci.aci import ACIModule, aci_argument_spec
from ansible.module_utils.basic import AnsibleModule


def main():
    argument_spec = aci_argument_spec
    argument_spec.update(
        leaf_profile=dict(type='str', aliases=['leaf_profile_name']),
        interface_selector=dict(type='str', aliases=['name', 'interface_selector_name', 'interface_profile_name']),
        state=dict(type='str', default='present', choices=['absent', 'present', 'query'])
    )

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
        required_if=[
            ['state', 'absent', ['leaf_profile', 'interface_selector']],
            ['state', 'present', ['leaf_profile', 'interface_selector']]
        ],
    )

    leaf_profile = module.params['leaf_profile']
    # WARNING: interface_selector accepts non existing interface_profile names and they appear on APIC gui with a state of "missing-target"
    interface_selector = module.params['interface_selector']
    state = module.params['state']

    # Defining the interface profile tDn for clarity
    interface_selector_tDn = 'uni/infra/accportprof-{0}'.format(interface_selector)

    aci = ACIModule(module)
    aci.construct_url(
        root_class=dict(
            aci_class='infraNodeP',
            aci_rn='infra/nprof-{0}'.format(leaf_profile),
            filter_target='eq(infraNodeP.name, "{0}")'.format(leaf_profile),
            module_object=leaf_profile
        ),
        subclass_1=dict(
            aci_class='infraRsAccPortP',
            aci_rn='rsaccPortP-[{0}]'.format(interface_selector_tDn),
            filter_target='eq(infraRsAccPortP.name, "{0}")'.format(interface_selector),
            module_object=interface_selector,
        )
    )

    aci.get_existing()

    if state == 'present':
        # Filter out module params with null values
        aci.payload(
            aci_class='infraRsAccPortP',
            class_config=dict(tDn=interface_selector_tDn)
        )

        # Generate config diff which will be used as POST request body
        aci.get_diff(aci_class='infraRsAccPortP')

        # Submit changes if module not in check_mode and the proposed is different than existing
        aci.post_config()

    elif state == 'absent':
        aci.delete_config()

    module.exit_json(**aci.result)


if __name__ == "__main__":
    main()
