#!/usr/bin/python
# -*- coding: utf-8 -*-

# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function
__metaclass__ = type

ANSIBLE_METADATA = {'metadata_version': '1.1',
                    'status': ['preview'],
                    'supported_by': 'community'}

DOCUMENTATION = r'''
---
module: aci_domain
short_description: Manage physical, virtual, bridged, routed or FC domain profiles (*:DomP)
description:
- Manage physical, virtual, bridged, routed or FC domain profiles.
- More information from the internal APIC classes I(phys:DomP),
  I(vmm:DomP), I(l2ext:DomP), I(l3ext:DomP), I(fc:DomP) at
  U(https://developer.cisco.com/docs/apic-mim-ref/).
author:
- Dag Wieers (@dagwieers)
version_added: '2.5'
options:
  domain:
    description:
    - Name of the physical, virtual, bridged routed or FC domain profile.
    aliases: [ domain_name, domain_profile, name ]
  domain_type:
    description:
    - The type of domain profile.
    - 'C(fc): The FC domain profile is a policy pertaining to single FC Management domain'
    - 'C(l2dom): The external bridged domain profile is a policy for managing L2 bridged infrastructure bridged outside the fabric.'
    - 'C(l3dom): The external routed domain profile is a policy for managing L3 routed infrastructure outside the fabric.'
    - 'C(phys): The physical domain profile stores the physical resources and encap resources that should be used for EPGs associated with this domain.'
    - 'C(vmm): The VMM domain profile is a policy for grouping VM controllers with similar networking policy requirements.'
    choices: [ fc, l2dom, l3dom, phys, vmm ]
    aliases: [ type ]
  dscp:
    description:
    - The target Differentiated Service (DSCP) value.
    choices: [ AF11, AF12, AF13, AF21, AF22, AF23, AF31, AF32, AF33, AF41, AF42, AF43, CS0, CS1, CS2, CS3, CS4, CS5, CS6, CS7, EF, VA, unspecified ]
    default: unspecified
    aliases: [ target ]
  encap_mode:
    description:
    - The layer 2 encapsulation protocol to use with the virtual switch.
    choices: [ unknown, vlan, vxlan ]
  multicast_address:
    description:
    - The muticast IP address to use for the virtual switch.
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
  vswitch:
    description:
    - The virtual switch to use for vmm domains.
    choices: [ avs, default, dvs, unknown ]
    default: default
extends_documentation_fragment: aci
'''

EXAMPLES = r'''
- name: Add a new physical domain
  aci_domain:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    domain: phys_dom
    domain_type: phys
    state: present

- name: Remove a physical domain
  aci_domain:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    domain: phys_dom
    domain_type: phys
    state: absent

- name: Add a new VMM domain
  aci_domain:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    domain: hyperv_dom
    domain_type: vmm
    vm_provider: microsoft
    state: present

- name: Remove a VMM domain
  aci_domain:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    domain: hyperv_dom
    domain_type: vmm
    vm_provider: microsoft
    state: absent

- name: Query a specific physical domain
  aci_domain:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    domain: phys_dom
    domain_type: phys
    state: query

- name: Query all domains
  aci_domain:
    hostname: apic
    username: admin
    password: SomeSecretPassword
    state: query
'''

RETURN = r''' # '''

from ansible.module_utils.basic import AnsibleModule
from ansible.module_utils.network.aci.aci import ACIModule, aci_argument_spec

VM_PROVIDER_MAPPING = dict(
    microsoft="Microsoft",
    openstack="OpenStack",
    redhat="Redhat",
    vmware="VMware",
)
VSWITCH_MAPPING = dict(
    avs='n1kv',
    default='default',
    dvs='default',
    unknown='unknown',
)


def main():
    argument_spec = aci_argument_spec
    argument_spec.update(
        dscp=dict(type='str',
                  choices=['AF11', 'AF12', 'AF13', 'AF21', 'AF22', 'AF23', 'AF31', 'AF32', 'AF33', 'AF41', 'AF42', 'AF43',
                           'CS0', 'CS1', 'CS2', 'CS3', 'CS4', 'CS5', 'CS6', 'CS7', 'EF', 'VA', 'unspecified'],
                  aliases=['target']),
        domain=dict(type='str', aliases=['domain_name', 'domain_profile', 'name']),
        domain_type=dict(type='str', choices=['fc', 'l2dom', 'l3dom', 'phys', 'vmm'], aliases=['type']),
        encap_mode=dict(type='str', choices=['unknown', 'vlan', 'vxlan']),
        multicast_address=dict(type='str'),
        state=dict(type='str', default='present', choices=['absent', 'present', 'query']),
        vm_provider=dict(type='str', choices=['microsoft', 'openstack', 'redhat', 'vmware']),
        vswitch=dict(type='str', choices=['avs', 'default', 'dvs', 'unknown']),
    )

    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True,
        required_if=[
            ['domain_type', 'vmm', ['vm_provider']],
            ['state', 'absent', ['domain', 'domain_type']],
            ['state', 'present', ['domain', 'domain_type']],
        ],
    )

    dscp = module.params['dscp']
    domain = module.params['domain']
    domain_type = module.params['domain_type']
    encap_mode = module.params['encap_mode']
    multicast_address = module.params['multicast_address']
    vm_provider = module.params['vm_provider']
    vswitch = module.params['vswitch']
    if vswitch is not None:
        vswitch = VSWITCH_MAPPING[vswitch]
    state = module.params['state']

    if domain_type != 'vmm':
        if vm_provider is not None:
            module.fail_json(msg="Domain type '{0}' cannot have a 'vm_provider'".format(domain_type))
        if encap_mode is not None:
            module.fail_json(msg="Domain type '{0}' cannot have an 'encap_mode'".format(domain_type))
        if multicast_address is not None:
            module.fail_json(msg="Domain type '{0}' cannot have a 'multicast_address'".format(domain_type))
        if vswitch is not None:
            module.fail_json(msg="Domain type '{0}' cannot have a 'vswitch'".format(domain_type))

    if dscp is not None and domain_type not in ['l2dom', 'l3dom']:
        module.fail_json(msg="DSCP values can only be assigned to 'l2ext and 'l3ext' domains")

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

    aci = ACIModule(module)
    aci.construct_url(
        root_class=dict(
            aci_class=domain_class,
            aci_rn=domain_rn,
            filter_target='eq({0}.name, "{1}")'.format(domain_class, domain),
            module_object=domain_mo,
        ),
    )

    aci.get_existing()

    if state == 'present':
        # Filter out module parameters with null values
        aci.payload(
            aci_class=domain_class,
            class_config=dict(
                encapMode=encap_mode,
                mcastAddr=multicast_address,
                mode=vswitch,
                name=domain,
                targetDscp=dscp,
            ),
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
