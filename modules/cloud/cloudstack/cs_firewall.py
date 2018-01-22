#!/usr/bin/python
# -*- coding: utf-8 -*-
#
# Copyright: (c) 2015, René Moser <mail@renemoser.net>
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)


ANSIBLE_METADATA = {'metadata_version': '1.1',
                    'status': ['stableinterface'],
                    'supported_by': 'community'}


DOCUMENTATION = '''
---
module: cs_firewall
short_description: Manages firewall rules on Apache CloudStack based clouds.
description:
    - Creates and removes firewall rules.
version_added: '2.0'
author: "René Moser (@resmo)"
options:
  ip_address:
    description:
      - Public IP address the ingress rule is assigned to.
      - Required if C(type=ingress).
  network:
    description:
      - Network the egress rule is related to.
      - Required if C(type=egress).
  state:
    description:
      - State of the firewall rule.
    default: present
    choices: [ present, absent ]
  type:
    description:
      - Type of the firewall rule.
    default: ingress
    choices: [ ingress, egress ]
  protocol:
    description:
      - Protocol of the firewall rule.
      - C(all) is only available if C(type=egress).
    default: tcp
    choices: [ tcp, udp, icmp, all ]
  cidrs:
    description:
      - List of CIDRs (full notation) to be used for firewall rule.
      - Since version 2.5, it is a list of CIDR.
    default: 0.0.0.0/0
    aliases: [ cidr ]
  start_port:
    description:
      - Start port for this rule.
      - Considered if C(protocol=tcp) or C(protocol=udp).
    aliases: [ port ]
  end_port:
    description:
      - End port for this rule. Considered if C(protocol=tcp) or C(protocol=udp).
      - If not specified, equal C(start_port).
  icmp_type:
    description:
      - Type of the icmp message being sent.
      - Considered if C(protocol=icmp).
  icmp_code:
    description:
      - Error code for this icmp message.
      - Considered if C(protocol=icmp).
  domain:
    description:
      - Domain the firewall rule is related to.
  account:
    description:
      - Account the firewall rule is related to.
  project:
    description:
      - Name of the project the firewall rule is related to.
  zone:
    description:
      - Name of the zone in which the virtual machine is in.
      - If not set, default zone is used.
  poll_async:
    description:
      - Poll async jobs until job has finished.
    default: true
    type: bool
  tags:
    description:
      - List of tags. Tags are a list of dictionaries having keys C(key) and C(value).
      - "To delete all tags, set a empty list e.g. C(tags: [])."
    aliases: [ tag ]
    version_added: "2.4"
extends_documentation_fragment: cloudstack
'''

EXAMPLES = '''
- name: Allow inbound port 80/tcp from 1.2.3.4 to 4.3.2.1
  local_action:
    module: cs_firewall
    ip_address: 4.3.2.1
    port: 80
    cidr: 1.2.3.4/32

- name: Allow inbound tcp/udp port 53 to 4.3.2.1
  local_action:
    module: cs_firewall
    ip_address: 4.3.2.1
    port: 53
    protocol: '{{ item }}'
  with_items:
  - tcp
  - udp

- name: Ensure firewall rule is removed
  local_action:
    module: cs_firewall
    ip_address: 4.3.2.1
    start_port: 8000
    end_port: 8888
    cidr: 17.0.0.0/8
    state: absent

- name: Allow all outbound traffic
  local_action:
    module: cs_firewall
    network: my_network
    type: egress
    protocol: all

- name: Allow only HTTP outbound traffic for an IP
  local_action:
    module: cs_firewall
    network: my_network
    type: egress
    port: 80
    cidr: 10.101.1.20
'''

RETURN = '''
---
id:
  description: UUID of the rule.
  returned: success
  type: string
  sample: 04589590-ac63-4ffc-93f5-b698b8ac38b6
ip_address:
  description: IP address of the rule if C(type=ingress)
  returned: success
  type: string
  sample: 10.100.212.10
type:
  description: Type of the rule.
  returned: success
  type: string
  sample: ingress
cidr:
  description: CIDR string of the rule.
  returned: success
  type: string
  sample: 0.0.0.0/0
cidrs:
  description: CIDR list of the rule.
  returned: success
  type: list
  sample: [ '0.0.0.0/0' ]
  version_added: "2.5"
protocol:
  description: Protocol of the rule.
  returned: success
  type: string
  sample: tcp
start_port:
  description: Start port of the rule.
  returned: success
  type: int
  sample: 80
end_port:
  description: End port of the rule.
  returned: success
  type: int
  sample: 80
icmp_code:
  description: ICMP code of the rule.
  returned: success
  type: int
  sample: 1
icmp_type:
  description: ICMP type of the rule.
  returned: success
  type: int
  sample: 1
network:
  description: Name of the network if C(type=egress)
  returned: success
  type: string
  sample: my_network
'''

from ansible.module_utils.basic import AnsibleModule
from ansible.module_utils.cloudstack import (
    AnsibleCloudStack,
    cs_argument_spec,
    cs_required_together
)


class AnsibleCloudStackFirewall(AnsibleCloudStack):

    def __init__(self, module):
        super(AnsibleCloudStackFirewall, self).__init__(module)
        self.returns = {
            'cidrlist': 'cidr',
            'startport': 'start_port',
            'endport': 'end_port',
            'protocol': 'protocol',
            'ipaddress': 'ip_address',
            'icmpcode': 'icmp_code',
            'icmptype': 'icmp_type',
        }
        self.firewall_rule = None
        self.network = None

    def get_firewall_rule(self):
        if not self.firewall_rule:
            cidrs = self.module.params.get('cidrs')
            protocol = self.module.params.get('protocol')
            start_port = self.module.params.get('start_port')
            end_port = self.get_or_fallback('end_port', 'start_port')
            icmp_code = self.module.params.get('icmp_code')
            icmp_type = self.module.params.get('icmp_type')
            fw_type = self.module.params.get('type')

            if protocol in ['tcp', 'udp'] and not (start_port and end_port):
                self.module.fail_json(msg="missing required argument for protocol '%s': start_port or end_port" % protocol)

            if protocol == 'icmp' and not icmp_type:
                self.module.fail_json(msg="missing required argument for protocol 'icmp': icmp_type")

            if protocol == 'all' and fw_type != 'egress':
                self.module.fail_json(msg="protocol 'all' could only be used for type 'egress'")

            args = {
                'account': self.get_account('name'),
                'domainid': self.get_domain('id'),
                'projectid': self.get_project('id')
            }
            if fw_type == 'egress':
                args['networkid'] = self.get_network(key='id')
                if not args['networkid']:
                    self.module.fail_json(msg="missing required argument for type egress: network")
                firewall_rules = self.query_api('listEgressFirewallRules', **args)
            else:
                args['ipaddressid'] = self.get_ip_address('id')
                if not args['ipaddressid']:
                    self.module.fail_json(msg="missing required argument for type ingress: ip_address")
                firewall_rules = self.query_api('listFirewallRules', **args)

            if firewall_rules and 'firewallrule' in firewall_rules:
                for rule in firewall_rules['firewallrule']:
                    type_match = self._type_cidrs_match(rule, cidrs)

                    protocol_match = (
                        self._tcp_udp_match(rule, protocol, start_port, end_port) or
                        self._icmp_match(rule, protocol, icmp_code, icmp_type) or
                        self._egress_all_match(rule, protocol, fw_type)
                    )

                    if type_match and protocol_match:
                        self.firewall_rule = rule
                        break
        return self.firewall_rule

    def _tcp_udp_match(self, rule, protocol, start_port, end_port):
        return (
            protocol in ['tcp', 'udp'] and
            protocol == rule['protocol'] and
            start_port == int(rule['startport']) and
            end_port == int(rule['endport'])
        )

    def _egress_all_match(self, rule, protocol, fw_type):
        return (
            protocol in ['all'] and
            protocol == rule['protocol'] and
            fw_type == 'egress'
        )

    def _icmp_match(self, rule, protocol, icmp_code, icmp_type):
        return (
            protocol == 'icmp' and
            protocol == rule['protocol'] and
            icmp_code == rule['icmpcode'] and
            icmp_type == rule['icmptype']
        )

    def _type_cidrs_match(self, rule, cidrs):
        return ",".join(cidrs) == rule['cidrlist']

    def create_firewall_rule(self):
        firewall_rule = self.get_firewall_rule()
        if not firewall_rule:
            self.result['changed'] = True

            args = {
                'cidrlist': self.module.params.get('cidrs'),
                'protocol': self.module.params.get('protocol'),
                'startport': self.module.params.get('start_port'),
                'endport': self.get_or_fallback('end_port', 'start_port'),
                'icmptype': self.module.params.get('icmp_type'),
                'icmpcode': self.module.params.get('icmp_code')
            }

            fw_type = self.module.params.get('type')
            if not self.module.check_mode:
                if fw_type == 'egress':
                    args['networkid'] = self.get_network(key='id')
                    res = self.query_api('createEgressFirewallRule', **args)
                else:
                    args['ipaddressid'] = self.get_ip_address('id')
                    res = self.query_api('createFirewallRule', **args)

                poll_async = self.module.params.get('poll_async')
                if poll_async:
                    firewall_rule = self.poll_job(res, 'firewallrule')

        if firewall_rule:
            firewall_rule = self.ensure_tags(resource=firewall_rule, resource_type='Firewallrule')
            self.firewall_rule = firewall_rule

        return firewall_rule

    def remove_firewall_rule(self):
        firewall_rule = self.get_firewall_rule()
        if firewall_rule:
            self.result['changed'] = True

            args = {
                'id': firewall_rule['id']
            }

            fw_type = self.module.params.get('type')
            if not self.module.check_mode:
                if fw_type == 'egress':
                    res = self.query_api('deleteEgressFirewallRule', **args)
                else:
                    res = self.query_api('deleteFirewallRule', **args)

                poll_async = self.module.params.get('poll_async')
                if poll_async:
                    self.poll_job(res, 'firewallrule')
        return firewall_rule

    def get_result(self, firewall_rule):
        super(AnsibleCloudStackFirewall, self).get_result(firewall_rule)
        if firewall_rule:
            self.result['type'] = self.module.params.get('type')
            if self.result['type'] == 'egress':
                self.result['network'] = self.get_network(key='displaytext')
            if 'cidrlist' in firewall_rule:
                self.result['cidrs'] = firewall_rule['cidrlist'].split(',') or [firewall_rule['cidrlist']]
        return self.result


def main():
    argument_spec = cs_argument_spec()
    argument_spec.update(dict(
        ip_address=dict(),
        network=dict(),
        cidrs=dict(type='list', default='0.0.0.0/0', aliases=['cidr']),
        protocol=dict(choices=['tcp', 'udp', 'icmp', 'all'], default='tcp'),
        type=dict(choices=['ingress', 'egress'], default='ingress'),
        icmp_type=dict(type='int'),
        icmp_code=dict(type='int'),
        start_port=dict(type='int', aliases=['port']),
        end_port=dict(type='int'),
        state=dict(choices=['present', 'absent'], default='present'),
        zone=dict(),
        domain=dict(),
        account=dict(),
        project=dict(),
        poll_async=dict(type='bool', default=True),
        tags=dict(type='list', aliases=['tag'], default=None),
    ))

    required_together = cs_required_together()
    required_together.extend([
        ['icmp_type', 'icmp_code'],
    ])

    module = AnsibleModule(
        argument_spec=argument_spec,
        required_together=required_together,
        required_one_of=(
            ['ip_address', 'network'],
        ),
        mutually_exclusive=(
            ['icmp_type', 'start_port'],
            ['icmp_type', 'end_port'],
            ['ip_address', 'network'],
        ),
        supports_check_mode=True
    )

    acs_fw = AnsibleCloudStackFirewall(module)

    state = module.params.get('state')
    if state in ['absent']:
        fw_rule = acs_fw.remove_firewall_rule()
    else:
        fw_rule = acs_fw.create_firewall_rule()

    result = acs_fw.get_result(fw_rule)

    module.exit_json(**result)


if __name__ == '__main__':
    main()
