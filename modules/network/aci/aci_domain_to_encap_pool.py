#!/usr/bin/python
# -*- coding: utf-8 -*-

# Copyright: (c) 2017, Dag Wieers <dag@wieers.com>
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function
__metaclass__ = type

ANSIBLE_METADATA = {'metadata_version': '1.1',
                    'status': ['preview'],
                    'supported_by': 'community'}

DOCUMENTATION = r'''
---
module: aci_domain_to_encap_pool
short_description: Bind Domain to Encap Pools on Cisco ACI fabrics (infra:RsVlanNs)
description:
- Bind Domain to Encap Pools on Cisco ACI fabrics.
- More information from the internal APIC class I(infra:RsVlanNs) at
  U(https://developer.cisco.com/docs/apic-mim-ref/).
author:
- Dag Wieers (@dagwieers)
version_added: '2.5'
notes:
- The C(domain) and C(encap_pool) parameters should exist before using this module.
  The M(aci_domain) and M(aci_encap_pool) can be used for these.
options:
  domain:
    description:
    - Name of the domain being associated with the Encap Pool.
    aliases: [ domain_name, domain_profile ]
  domain_type:
    description:
    - Determines if the Domain is physical (phys) or virtual (vmm).
    choices: [ fc, l2dom, l3dom, phys, vmm ]
  pool:
    description:
    - The name of the pool.
    aliases: [ pool_name ]
  pool_allocation_mode:
    description:
    - The method used for allocating encaps to resources.
    - Only vlan and vsan support allocation modes.
    aliases: [ mode ]
    choices: [ dynamic, static]
  pool_type:
    description:
    - The encap type of C(pool).
    required: yes
    choices: [ vlan, vsan, vxsan ]
  state:
    description:
    - Use C(present) or C(absent) for adding or removing.
    - Use C(query) for listing an object or multiple objects.
    choices: [ absent, present, query ]
    default: present
  vm_provider:
    description:
    - The VM platform for VMM Domains.
    choices: [ microsoft, openstack, redhat, vmware ]
'''

EXAMPLES = r''' # '''

RETURN = ''' # '''

from ansible.module_utils.network.aci.aci import ACIModule, aci_argument_spec
from ansible.module_utils.basic import AnsibleModule

VM_PROVIDER_MAPPING = dict(
    microsoft='Microsoft',
    openstack='OpenStack',
    redhat='Redhat',
    vmware='VMware',
)

POOL_MAPPING = dict(
    vlan=dict(
        aci_mo='uni/infra/vlanns-',
    ),
    vxlan=dict(
        aci_mo='uni/infra/vxlanns-',
    ),
    vsan=dict(
        aci_mo='uni/infra/vsanns-',
    ),
)


def main():
    argument_spec = aci_argument_spec
    argument_spec.update(
        domain=dict(type='str', aliases=['domain_name', 'domain_profile']),
        domain_type=dict(type='str', choices=['fc', 'l2dom', 'l3dom', 'phys', 'vmm']),
        pool=dict(type='str', aliases=['pool_name']),
        pool_allocation_mode=dict(type='str', aliases=['allocation_mode', 'mode'], choices=['dynamic', 'static']),
        pool_type=dict(type='str', required=True, choices=['vlan', 'vsan', 'vxlan']),
        state=dict(type='str', default='present', choices=['absent', 'present', 'query']),
        vm_provider=dict(type='str', choices=['microsoft', 'openstack', 'redhat', 'vmware']),
    )

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
        required_if=[
            ['domain_type', 'vmm', ['vm_provider']],
            ['state', 'absent', ['domain', 'domain_type', 'pool', 'pool_type']],
            ['state', 'present', ['domain', 'domain_type', 'pool', 'pool_type']],
        ],
    )

    domain = module.params['domain']
    domain_type = module.params['domain_type']
    pool = module.params['pool']
    pool_allocation_mode = module.params['pool_allocation_mode']
    pool_type = module.params['pool_type']
    vm_provider = module.params['vm_provider']
    state = module.params['state']

    # Report when vm_provider is set when type is not virtual
    if domain_type != 'vmm' and vm_provider is not None:
        module.fail_json(msg="Domain type '{0}' cannot have a 'vm_provider'".format(domain_type))

    # ACI Pool URL requires the allocation mode for vlan and vsan pools (ex: uni/infra/vlanns-[poolname]-static)
    pool_name = pool
    if pool_type != 'vxlan' and pool is not None:
        if pool_allocation_mode is not None:
            pool_name = '[{0}]-{1}'.format(pool, pool_allocation_mode)
        else:
            module.fail_json(msg="ACI requires the 'pool_allocation_mode' for 'pool_type' of 'vlan' and 'vsan' when 'pool' is provided")

    # Vxlan pools do not support allocation modes
    if pool_type == 'vxlan' and pool_allocation_mode is not None:
        module.fail_json(msg='vxlan pools do not support setting the allocation_mode; please remove this parameter from the task')

    # Compile the full domain for URL building
    if domain_type == 'fc':
        domain_class = 'fcDomP'
        domain_mo = 'uni/fc-{0}'.format(domain)
        domain_rn = 'fc-{0}'.format(domain)
    elif domain_type == 'l2ext':
        domain_class = 'l2extDomP'
        domain_mo = 'uni/l2dom-{0}'.format(domain)
        domain_rn = 'l2dom-{0}'.format(domain)
    elif domain_type == 'l3ext':
        domain_class = 'l3extDomP'
        domain_mo = 'uni/l3dom-{0}'.format(domain)
        domain_rn = 'l3dom-{0}'.format(domain)
    elif domain_type == 'phys':
        domain_class = 'physDomP'
        domain_mo = 'uni/phys-{0}'.format(domain)
        domain_rn = 'phys-{0}'.format(domain)
    elif domain_type == 'vmm':
        domain_class = 'vmmDomP'
        domain_mo = 'uni/vmmp-{0}/dom-{1}'.format(VM_PROVIDER_MAPPING[vm_provider], domain)
        domain_rn = 'vmmp-{0}/dom-{1}'.format(VM_PROVIDER_MAPPING[vm_provider], domain)

    pool_mo = POOL_MAPPING[pool_type]["aci_mo"] + pool_name

    aci = ACIModule(module)
    aci.construct_url(
        root_class=dict(
            aci_class=domain_class,
            aci_rn=domain_rn,
            filter_target='eq({0}.name, "{1}")'.format(domain_class, domain),
            module_object=domain_mo,
        ),
        child_classes=['infraRsVlanNs'],
    )

    aci.get_existing()

    if state == 'present':
        # Filter out module params with null values
        aci.payload(
            aci_class=domain_class,
            class_config=dict(name=domain),
            child_configs=[
                {'infraRsVlanNs': {'attributes': {'tDn': pool_mo}}},
            ]
        )

        # Generate config diff which will be used as POST request body
        aci.get_diff(aci_class=domain_class)

        # Submit changes if module not in check_mode and the proposed is different than existing
        aci.post_config()

    elif state == 'absent':
        aci.delete_config()

    module.exit_json(**aci.result)


if __name__ == "__main__":
    main()
