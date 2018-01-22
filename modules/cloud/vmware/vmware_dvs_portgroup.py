#!/usr/bin/python
# -*- coding: utf-8 -*-

# (c) 2015, Joseph Callen <jcallen () csc.com>
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)
#

from __future__ import absolute_import, division, print_function
__metaclass__ = type


ANSIBLE_METADATA = {'metadata_version': '1.1',
                    'status': ['preview'],
                    'supported_by': 'community'}


DOCUMENTATION = '''
---
module: vmware_dvs_portgroup
short_description: Create or remove a Distributed vSwitch portgroup.
description:
    - Create or remove a Distributed vSwitch portgroup.
version_added: 2.0
author:
    - Joseph Callen (@jcpowermac)
    - Philippe Dellaert (@pdellaert) <philippe@dellaert.org>
notes:
    - Tested on vSphere 5.5
    - Tested on vSphere 6.5
requirements:
    - "python >= 2.6"
    - PyVmomi
options:
    portgroup_name:
        description:
            - The name of the portgroup that is to be created or deleted.
        required: True
    switch_name:
        description:
            - The name of the distributed vSwitch the port group should be created on.
        required: True
    vlan_id:
        description:
            - The VLAN ID that should be configured with the portgroup, use 0 for no VLAN.
            - 'If C(vlan_trunk) is configured to be I(true), this can be a range, example: 1-4094.'
        required: True
    num_ports:
        description:
            - The number of ports the portgroup should contain.
        required: True
    portgroup_type:
        description:
            - See VMware KB 1022312 regarding portgroup types.
        required: True
        choices:
            - 'earlyBinding'
            - 'lateBinding'
            - 'ephemeral'
    state:
        description:
            - Determines if the portgroup should be present or not.
        required: True
        choices:
            - 'present'
            - 'absent'
        version_added: '2.5'
    vlan_trunk:
        description:
            - Indicates whether this is a VLAN trunk or not.
        required: False
        default: False
        version_added: '2.5'
    network_policy:
        description:
            - Dict which configures the different security values for portgroup.
            - 'Valid attributes are:'
            - '- C(promiscuous) (bool): indicates whether promiscuous mode is allowed. (default: false)'
            - '- C(forged_transmits) (bool): indicates whether forged transmits are allowed. (default: false)'
            - '- C(mac_changes) (bool): indicates whether mac changes are allowed. (default: false)'
        required: False
        version_added: '2.5'
    port_policy:
        description:
            - Dict which configures the advanced policy settings for the portgroup.
            - 'Valid attributes are:'
            - '- C(block_override) (bool): indicates if the block policy can be changed per port. (default: true)'
            - '- C(ipfix_override) (bool): indicates if the ipfix policy can be changed per port. (default: false)'
            - '- C(live_port_move) (bool): indicates if a live port can be moved in or out of the portgroup. (default: false)'
            - '- C(network_rp_override) (bool): indicates if the network resource pool can be changed per port. (default: false)'
            - '- C(port_config_reset_at_disconnect) (bool): indicates if the configuration of a port is reset automatically after disconnect. (default: true)'
            - '- C(security_override) (bool): indicates if the security policy can be changed per port. (default: false)'
            - '- C(shaping_override) (bool): indicates if the shaping policy can be changed per port. (default: false)'
            - '- C(traffic_filter_override) (bool): indicates if the traffic filter can be changed per port. (default: false)'
            - '- C(uplink_teaming_override) (bool): indicates if the uplink teaming policy can be changed per port. (default: false)'
            - '- C(vendor_config_override) (bool): indicates if the vendor config can be changed per port. (default: false)'
            - '- C(vlan_override) (bool): indicates if the vlan can be changed per port. (default: false)'
        required: False
        version_added: '2.5'
extends_documentation_fragment: vmware.documentation
'''

EXAMPLES = '''
   - name: Create vlan portgroup
     connection: local
     vmware_dvs_portgroup:
        hostname: vcenter_ip_or_hostname
        username: vcenter_username
        password: vcenter_password
        portgroup_name: vlan-123-portrgoup
        switch_name: dvSwitch
        vlan_id: 123
        num_ports: 120
        portgroup_type: earlyBinding
        state: present

   - name: Create vlan trunk portgroup
     connection: local
     vmware_dvs_portgroup:
        hostname: vcenter_ip_or_hostname
        username: vcenter_username
        password: vcenter_password
        portgroup_name: vlan-trunk-portrgoup
        switch_name: dvSwitch
        vlan_id: 1-1000
        vlan_trunk: True
        num_ports: 120
        portgroup_type: earlyBinding
        state: present

   - name: Create no-vlan portgroup
     connection: local
     vmware_dvs_portgroup:
        hostname: vcenter_ip_or_hostname
        username: vcenter_username
        password: vcenter_password
        portgroup_name: no-vlan-portrgoup
        switch_name: dvSwitch
        vlan_id: 0
        num_ports: 120
        portgroup_type: earlyBinding
        state: present

   - name: Create vlan portgroup with all security and port policies
     connection: local
     vmware_dvs_portgroup:
        hostname: vcenter_ip_or_hostname
        username: vcenter_username
        password: vcenter_password
        portgroup_name: vlan-123-portrgoup
        switch_name: dvSwitch
        vlan_id: 123
        num_ports: 120
        portgroup_type: earlyBinding
        state: present
        network_policy:
          promiscuous: yes
          forged_transmits: yes
          mac_changes: yes
        port_policy:
          block_override: yes
          ipfix_override: yes
          live_port_move: yes
          network_rp_override: yes
          port_config_reset_at_disconnect: yes
          security_override: yes
          shaping_override: yes
          traffic_filter_override: yes
          uplink_teaming_override: yes
          vendor_config_override: yes
          vlan_override: yes
'''

try:
    from pyVmomi import vim, vmodl
    HAS_PYVMOMI = True
except ImportError:
    HAS_PYVMOMI = False

from ansible.module_utils.basic import AnsibleModule
from ansible.module_utils.vmware import (HAS_PYVMOMI, connect_to_api, find_dvs_by_name, find_dvspg_by_name,
                                         vmware_argument_spec, wait_for_task)


class VMwareDvsPortgroup(object):
    def __init__(self, module):
        self.module = module
        self.dvs_portgroup = None
        self.switch_name = self.module.params['switch_name']
        self.portgroup_name = self.module.params['portgroup_name']
        self.vlan_id = self.module.params['vlan_id']
        self.num_ports = self.module.params['num_ports']
        self.portgroup_type = self.module.params['portgroup_type']
        self.dv_switch = None
        self.state = self.module.params['state']
        self.vlan_trunk = self.module.params['vlan_trunk']
        self.security_promiscuous = self.module.params['network_policy']['promiscuous']
        self.security_forged_transmits = self.module.params['network_policy']['forged_transmits']
        self.security_mac_changes = self.module.params['network_policy']['mac_changes']
        self.policy_block_override = self.module.params['port_policy']['block_override']
        self.policy_ipfix_override = self.module.params['port_policy']['ipfix_override']
        self.policy_live_port_move = self.module.params['port_policy']['live_port_move']
        self.policy_network_rp_override = self.module.params['port_policy']['network_rp_override']
        self.policy_port_config_reset_at_disconnect = self.module.params['port_policy']['port_config_reset_at_disconnect']
        self.policy_security_override = self.module.params['port_policy']['security_override']
        self.policy_shaping_override = self.module.params['port_policy']['shaping_override']
        self.policy_traffic_filter_override = self.module.params['port_policy']['traffic_filter_override']
        self.policy_uplink_teaming_override = self.module.params['port_policy']['uplink_teaming_override']
        self.policy_vendor_config_override = self.module.params['port_policy']['vendor_config_override']
        self.policy_vlan_override = self.module.params['port_policy']['vlan_override']
        self.content = connect_to_api(module)

    def process_state(self):
        try:
            dvspg_states = {
                'absent': {
                    'present': self.state_destroy_dvspg,
                    'absent': self.state_exit_unchanged,
                },
                'present': {
                    'update': self.state_update_dvspg,
                    'present': self.state_exit_unchanged,
                    'absent': self.state_create_dvspg,
                }
            }
            dvspg_states[self.state][self.check_dvspg_state()]()
        except vmodl.RuntimeFault as runtime_fault:
            self.module.fail_json(msg=runtime_fault.msg)
        except vmodl.MethodFault as method_fault:
            self.module.fail_json(msg=method_fault.msg)
        except Exception as e:
            self.module.fail_json(msg=str(e))

    def create_port_group(self):
        config = vim.dvs.DistributedVirtualPortgroup.ConfigSpec()

        # Basic config
        config.name = self.portgroup_name
        config.numPorts = self.num_ports

        # Default port config
        config.defaultPortConfig = vim.dvs.VmwareDistributedVirtualSwitch.VmwarePortConfigPolicy()
        if self.vlan_trunk:
            config.defaultPortConfig.vlan = vim.dvs.VmwareDistributedVirtualSwitch.TrunkVlanSpec()
            vlan_id_start, vlan_id_end = self.vlan_id.split('-')
            config.defaultPortConfig.vlan.vlanId = [vim.NumericRange(start=int(vlan_id_start.strip()), end=int(vlan_id_end.strip()))]
        else:
            config.defaultPortConfig.vlan = vim.dvs.VmwareDistributedVirtualSwitch.VlanIdSpec()
            config.defaultPortConfig.vlan.vlanId = int(self.vlan_id)
        config.defaultPortConfig.vlan.inherited = False
        config.defaultPortConfig.securityPolicy = vim.dvs.VmwareDistributedVirtualSwitch.SecurityPolicy()
        config.defaultPortConfig.securityPolicy.allowPromiscuous = vim.BoolPolicy(value=self.security_promiscuous)
        config.defaultPortConfig.securityPolicy.forgedTransmits = vim.BoolPolicy(value=self.security_forged_transmits)
        config.defaultPortConfig.securityPolicy.macChanges = vim.BoolPolicy(value=self.security_mac_changes)

        # PG policy (advanced_policy)
        config.policy = vim.dvs.VmwareDistributedVirtualSwitch.VMwarePortgroupPolicy()
        config.policy.blockOverrideAllowed = self.policy_block_override
        config.policy.ipfixOverrideAllowed = self.policy_ipfix_override
        config.policy.livePortMovingAllowed = self.policy_live_port_move
        config.policy.networkResourcePoolOverrideAllowed = self.policy_network_rp_override
        config.policy.portConfigResetAtDisconnect = self.policy_port_config_reset_at_disconnect
        config.policy.securityPolicyOverrideAllowed = self.policy_security_override
        config.policy.shapingOverrideAllowed = self.policy_shaping_override
        config.policy.trafficFilterOverrideAllowed = self.policy_traffic_filter_override
        config.policy.uplinkTeamingOverrideAllowed = self.policy_uplink_teaming_override
        config.policy.vendorConfigOverrideAllowed = self.policy_vendor_config_override
        config.policy.vlanOverrideAllowed = self.policy_vlan_override

        # PG Type
        config.type = self.portgroup_type

        spec = [config]
        task = self.dv_switch.AddDVPortgroup_Task(spec)
        changed, result = wait_for_task(task)
        return changed, result

    def state_destroy_dvspg(self):
        changed = True
        result = None

        if not self.module.check_mode:
            task = self.dvs_portgroup.Destroy_Task()
            changed, result = wait_for_task(task)
        self.module.exit_json(changed=changed, result=str(result))

    def state_exit_unchanged(self):
        self.module.exit_json(changed=False)

    def state_update_dvspg(self):
        self.module.exit_json(changed=False, msg="Currently not implemented.")

    def state_create_dvspg(self):
        changed = True
        result = None

        if not self.module.check_mode:
            changed, result = self.create_port_group()
        self.module.exit_json(changed=changed, result=str(result))

    def check_dvspg_state(self):
        self.dv_switch = find_dvs_by_name(self.content, self.switch_name)

        if self.dv_switch is None:
            raise Exception("A distributed virtual switch with name %s does not exist" % self.switch_name)
        self.dvs_portgroup = find_dvspg_by_name(self.dv_switch, self.portgroup_name)

        if self.dvs_portgroup is None:
            return 'absent'
        else:
            return 'present'


def main():
    argument_spec = vmware_argument_spec()
    argument_spec.update(
        dict(
            portgroup_name=dict(required=True, type='str'),
            switch_name=dict(required=True, type='str'),
            vlan_id=dict(required=True, type='str'),
            num_ports=dict(required=True, type='int'),
            portgroup_type=dict(required=True, choices=['earlyBinding', 'lateBinding', 'ephemeral'], type='str'),
            state=dict(required=True, choices=['present', 'absent'], type='str'),
            vlan_trunk=dict(type='bool', default=False),
            network_policy=dict(
                type='dict',
                options=dict(
                    promiscuous=dict(type='bool', default=False),
                    forged_transmits=dict(type='bool', default=False),
                    mac_changes=dict(type='bool', default=False)
                ),
                default=dict(
                    promiscuous=False,
                    forged_transmits=False,
                    mac_changes=False
                )
            ),
            port_policy=dict(
                type='dict',
                options=dict(
                    block_override=dict(type='bool', default=True),
                    ipfix_override=dict(type='bool', default=False),
                    live_port_move=dict(type='bool', default=False),
                    network_rp_override=dict(type='bool', default=False),
                    port_config_reset_at_disconnect=dict(type='bool', default=True),
                    security_override=dict(type='bool', default=False),
                    shaping_override=dict(type='bool', default=False),
                    traffic_filter_override=dict(type='bool', default=False),
                    uplink_teaming_override=dict(type='bool', default=False),
                    vendor_config_override=dict(type='bool', default=False),
                    vlan_override=dict(type='bool', default=False)
                ),
                default=dict(
                    block_override=True,
                    ipfix_override=False,
                    live_port_move=False,
                    network_rp_override=False,
                    port_config_reset_at_disconnect=True,
                    security_override=False,
                    shaping_override=False,
                    traffic_filter_override=False,
                    uplink_teaming_override=False,
                    vendor_config_override=False,
                    vlan_override=False
                )
            )
        )
    )

    module = AnsibleModule(argument_spec=argument_spec, supports_check_mode=True)

    if not HAS_PYVMOMI:
        module.fail_json(msg='pyvmomi is required for this module')

    vmware_dvs_portgroup = VMwareDvsPortgroup(module)
    vmware_dvs_portgroup.process_state()


if __name__ == '__main__':
    main()
