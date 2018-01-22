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
module: aci_domain_to_vlan_pool
short_description: Bind Domain to VLAN Pools on Cisco ACI fabrics (infra:RsVlanNs)
description:
- Bind Domain to VLAN Pools on Cisco ACI fabrics.
- More information from the internal APIC class I(infra:RsVlanNs) at
  U(https://developer.cisco.com/docs/apic-mim-ref/).
author:
- Dag Wieers (@dagwieers)
version_added: '2.5'
notes:
- The C(domain) and C(vlan_pool) parameters should exist before using this module.
  The M(aci_domain) and M(aci_vlan_pool) can be used for these.
options:
  domain:
    description:
    - Name of the domain being associated with the VLAN Pool.
    aliases: [ domain_name, domain_profile ]
  domain_type:
    description:
    - Determines if the Domain is physical (phys) or virtual (vmm).
    choices: [ fc, l2dom, l3dom, phys, vmm ]
  pool:
    description:
    - The name of the pool.
    aliases: [ pool_name, vlan_pool ]
  pool_allocation_mode:
    description:
    - The method used for allocating VLANs to resources.
    aliases: [ mode ]
    choices: [ dynamic, static]
    required: yes
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

EXAMPLES = r'''
- name: Bind a VMM domain to VLAN pool
  aci_domain_to_vlan_pool:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    domain: vmw_dom
    domain_type: vmm
    pool: vmw_pool
    pool_allocation_mode: dynamic
    vm_provider: vmware
    state: present

- name: Remove a VMM domain to VLAN pool binding
  aci_domain_to_vlan_pool:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    domain: vmw_dom
    domain_type: vmm
    pool: vmw_pool
    pool_allocation_mode: dynamic
    vm_provider: vmware
    state: absent

- name: Bind a physical domain to VLAN pool
  aci_domain_to_vlan_pool:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    domain: phys_dom
    domain_type: phys
    pool: phys_pool
    pool_allocation_mode: static
    state: present

- name: Bind a physical domain to VLAN pool
  aci_domain_to_vlan_pool:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    domain: phys_dom
    domain_type: phys
    pool: phys_pool
    pool_allocation_mode: static
    state: absent

- name: Query an domain to VLAN pool binding
  aci_domain_to_vlan_pool:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    domain: phys_dom
    domain_type: phys
    pool: phys_pool
    pool_allocation_mode: static
    state: query

- name: Query all domain to VLAN pool bindings
  aci_domain_to_vlan_pool:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    state: query
'''

RETURN = ''' # '''

from ansible.module_utils.network.aci.aci import ACIModule, aci_argument_spec
from ansible.module_utils.basic import AnsibleModule

VM_PROVIDER_MAPPING = dict(
    microsoft='Microsoft',
    openstack='OpenStack',
    redhat='Redhat',
    vmware='VMware',
)


def main():
    argument_spec = aci_argument_spec
    argument_spec.update(
        domain=dict(type='str', aliases=['domain_name', 'domain_profile']),
        domain_type=dict(type='str', choices=['fc', 'l2dom', 'l3dom', 'phys', 'vmm']),
        pool=dict(type='str', aliases=['pool_name']),
        pool_allocation_mode=dict(type='str', required=True, aliases=['allocation_mode', 'mode'], choices=['dynamic', 'static']),
        state=dict(type='str', default='present', choices=['absent', 'present', 'query']),
        vm_provider=dict(type='str', choices=['microsoft', 'openstack', 'redhat', 'vmware']),
    )

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
        required_if=[
            ['domain_type', 'vmm', ['vm_provider']],
            ['state', 'absent', ['domain', 'domain_type', 'pool']],
            ['state', 'present', ['domain', 'domain_type', 'pool']],
        ],
    )

    domain = module.params['domain']
    domain_type = module.params['domain_type']
    pool = module.params['pool']
    pool_allocation_mode = module.params['pool_allocation_mode']
    vm_provider = module.params['vm_provider']
    state = module.params['state']

    # Report when vm_provider is set when type is not virtual
    if domain_type != 'vmm' and vm_provider is not None:
        module.fail_json(msg="Domain type '{0}' cannot have a 'vm_provider'".format(domain_type))

    # ACI Pool URL requires the allocation mode for vlan and vsan pools (ex: uni/infra/vlanns-[poolname]-static)
    pool_name = pool
    if pool is not None:
        pool_name = '[{0}]-{1}'.format(pool, pool_allocation_mode)

    # Compile the full domain for URL building
    if domain_type == 'fc':
        domain_class = 'fcDomP'
        domain_mo = 'uni/fc-{0}'.format(domain)
        domain_rn = 'fc-{0}'.format(domain)
    elif domain_type == 'l2dom':
        domain_class = 'l2extDomP'
        domain_mo = 'uni/l2dom-{0}'.format(domain)
        domain_rn = 'l2dom-{0}'.format(domain)
    elif domain_type == 'l3dom':
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
        domain_rn = 'dom-{0}'.format(domain)

    aci_mo = 'uni/infra/vlanns-' + pool_name

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
            class_config=dict(name=domain_mo),
            child_configs=[
                {'infraRsVlanNs': {'attributes': {'tDn': aci_mo}}},
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
