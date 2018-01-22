#!/usr/bin/python
# -*- coding: utf-8 -*-

# Copyright: (c) 2017, Jacob McGill (jmcgill298)
# Copyright: (c) 2018, Dag Wieers (dagwieers) <dag@wieers.com>
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function
__metaclass__ = type

ANSIBLE_METADATA = {'metadata_version': '1.1',
                    'status': ['preview'],
                    'supported_by': 'community'}

DOCUMENTATION = r'''
---
module: aci_vlan_pool_encap_block
short_description: Manage encap blocks assigned to VLAN pools on Cisco ACI fabrics (fvns:EncapBlk)
description:
- Manage VLAN encap blocks that are assigned to VLAN pools on Cisco ACI fabrics.
- More information from the internal APIC class I(fvns:EncapBlk) at
  U(https://developer.cisco.com/docs/apic-mim-ref/).
author:
- Jacob McGill (@jmcgill298)
- Dag Wieers (@dagwieers)
version_added: '2.5'
requirements:
- The C(pool) must exist in order to add or delete a encap block.
options:
  allocation_mode:
    description:
    - The method used for allocating encaps to resources.
    aliases: [ mode ]
    choices: [ dynamic, inherit, static]
  description:
    description:
    - Description for the pool encap block.
    aliases: [ descr ]
  pool:
    description:
    - The name of the pool that the encap block should be assigned to.
    aliases: [ pool_name ]
  block_end:
    description:
    - The end of encap block.
    aliases: [ end ]
  block_name:
    description:
    - The name to give to the encap block.
    aliases: [ name, range ]
  block_start:
    description:
    - The start of the encap block.
    aliases: [ start ]
  state:
    description:
    - Use C(present) or C(absent) for adding or removing.
    - Use C(query) for listing an object or multiple objects.
    choices: [ absent, present, query ]
    default: present
extends_documentation_fragment: aci
'''

EXAMPLES = r'''
- name: Add a new VLAN encap block
  aci_vlan_pool_encap_block:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    pool: production
    block_start: 20
    block_end: 50
    state: present

- name: Remove a VLAN encap block
  aci_vlan_pool_encap_block:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    pool: production
    block_start: 20
    block_end: 50
    state: absent

- name: Query a VLAN encap block
  aci_vlan_pool_encap_block:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    pool: production
    block_start: 20
    block_end: 50
    state: query

- name: Query a VLAN pool for encap blocks
  aci_vlan_pool_encap_block:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    pool: production
    state: query

- name: Query all VLAN encap blocks
  aci_vlan_pool_encap_block:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    state: query
'''

RETURN = r'''
#
'''

from ansible.module_utils.network.aci.aci import ACIModule, aci_argument_spec
from ansible.module_utils.basic import AnsibleModule


def main():
    argument_spec = aci_argument_spec
    argument_spec.update(
        allocation_mode=dict(type='str', aliases=['mode'], choices=['dynamic', 'inherit', 'static']),
        description=dict(type='str', aliases=['descr']),
        pool=dict(type='str', aliases=['pool_name']),
        pool_allocation_mode=dict(type='str', aliases=['pool_mode'], choices=['dynamic', 'static']),
        block_name=dict(type='str', aliases=["name"]),
        block_end=dict(type='int', aliases=['end']),
        block_start=dict(type='int', aliases=["start"]),
        state=dict(type='str', default='present', choices=['absent', 'present', 'query']),
    )

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
        required_if=[
            ['state', 'absent', ['pool', 'block_end', 'block_name', 'block_start']],
            ['state', 'present', ['pool', 'block_end', 'block_name', 'block_start']],
        ],
    )

    allocation_mode = module.params['allocation_mode']
    description = module.params['description']
    pool = module.params['pool']
    pool_allocation_mode = module.params['pool_allocation_mode']
    block_end = module.params['block_end']
    block_name = module.params['block_name']
    block_start = module.params['block_start']
    state = module.params['state']

    if block_end is not None:
        encap_end = 'vlan-{0}'.format(block_end)
    else:
        encap_end = None

    if block_start is not None:
        encap_start = 'vlan-{0}'.format(block_start)
    else:
        encap_start = None

    # Collect proper mo information
    aci_block_mo = 'from-[{0}]-to-[{1}]'.format(encap_start, encap_end)
    pool_name = pool

    # Validate block_end and block_start are valid for its respective encap type
    for encap_id in block_end, block_start:
        if encap_id is not None:
            if not 1 <= encap_id <= 4094:
                module.fail_json(msg="vlan pools must have 'block_start' and 'block_end' values between 1 and 4094")

    # Build proper proper filter_target based on block_start, block_end, and block_name
    if block_end is not None and block_start is not None:
        # Validate block_start is less than block_end
        if block_start > block_end:
            module.fail_json(msg="The 'block_start' must be less than or equal to the 'block_end'")

        if block_name is None:
            block_filter_target = 'and(eq({0}.from, "{1}"),eq({0}.to, "{2}"))'.format('fvnsEncapBlk', encap_start, encap_end)
        else:
            block_filter_target = 'and(eq({0}.from, "{1}"),eq({0}.to, "{2}"),eq({0}.name, "{3}"))'.format('fvnsEncapBlk', encap_start, encap_end, block_name)
    elif block_end is None and block_start is None:
        if block_name is None:
            # Reset range managed object to None for aci util to properly handle query
            aci_block_mo = None
            block_filter_target = ''
        else:
            block_filter_target = 'eq({0}.name, "{1}")'.format('fvnsEncapBlk', block_name)
    elif block_start is not None:
        if block_name is None:
            block_filter_target = 'eq({0}.from, "{1}")'.format('fvnsEncapBlk', encap_start)
        else:
            block_filter_target = 'and(eq({0}.from, "{1}"),eq({0}.name, "{2}"))'.format('fvnsEncapBlk', encap_start, block_name)
    else:
        if block_name is None:
            block_filter_target = 'eq({0}.to, "{1}")'.format('fvnsEncapBlk', encap_end)
        else:
            block_filter_target = 'and(eq({0}.to, "{1}"),eq({0}.name, "{2}"))'.format('fvnsEncapBlk', encap_end, block_name)

    # ACI Pool URL requires the allocation mode (ex: uni/infra/vlanns-[poolname]-static)
    if pool is not None:
        if pool_allocation_mode is not None:
            pool_name = '[{0}]-{1}'.format(pool, pool_allocation_mode)
        else:
            module.fail_json(msg="ACI requires the 'pool_allocation_mode' when 'pool' is provided")

    aci = ACIModule(module)
    aci.construct_url(
        root_class=dict(
            aci_class='fvnsVlanInstP',
            aci_rn='infra/vlanns-{0}'.format(pool_name),
            filter_target='eq(fvnsVlanInstP.name, "{0}")'.format(pool),
            module_object=pool,
        ),
        subclass_1=dict(
            aci_class='fvnsEncapBlk',
            aci_rn=aci_block_mo,
            filter_target=block_filter_target,
            module_object=aci_block_mo,
        ),
    )

    aci.get_existing()

    if state == 'present':
        # Filter out module parameters with null values
        aci.payload(
            aci_class='fvnsEncapBlk',
            class_config={
                "allocMode": allocation_mode,
                "descr": description,
                "from": encap_start,
                "name": block_name,
                "to": encap_end,
            }
        )

        # Generate config diff which will be used as POST request body
        aci.get_diff(aci_class='fvnsEncapBlk')

        # Submit changes if module not in check_mode and the proposed is different than existing
        aci.post_config()

    elif state == 'absent':
        aci.delete_config()

    module.exit_json(**aci.result)


if __name__ == "__main__":
    main()
