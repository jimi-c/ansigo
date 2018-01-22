#!/usr/bin/python
# Copyright: Ansible Project
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function
__metaclass__ = type


ANSIBLE_METADATA = {'metadata_version': '1.1',
                    'status': ['stableinterface'],
                    'supported_by': 'core'}


DOCUMENTATION = '''
---
module: ec2_vpc_subnet
short_description: Manage subnets in AWS virtual private clouds
description:
    - Manage subnets in AWS virtual private clouds
version_added: "2.0"
author: Robert Estelle (@erydo), Brad Davidson (@brandond)
requirements: [ boto3 ]
options:
  az:
    description:
      - "The availability zone for the subnet."
    required: false
    default: null
  cidr:
    description:
      - "The CIDR block for the subnet. E.g. 192.0.2.0/24."
    required: false
    default: null
  ipv6_cidr:
    description:
      - "The IPv6 CIDR block for the subnet. The VPC must have a /56 block assigned and this value must be a valid IPv6 /64 that falls in the VPC range."
      - "Required if I(assign_instances_ipv6=true)"
    required: false
    default: null
    version_added: "2.5"
  tags:
    description:
      - "A dict of tags to apply to the subnet. Any tags currently applied to the subnet and not present here will be removed."
    required: false
    default: null
    aliases: [ 'resource_tags' ]
  state:
    description:
      - "Create or remove the subnet"
    required: false
    default: present
    choices: [ 'present', 'absent' ]
  vpc_id:
    description:
      - "VPC ID of the VPC in which to create or delete the subnet."
    required: true
    default: null
  map_public:
    description:
      - "Specify true to indicate that instances launched into the subnet should be assigned public IP address by default."
    required: false
    default: false
    version_added: "2.4"
  assign_instances_ipv6:
    description:
      - "Specify true to indicate that instances launched into the subnet should be automatically assigned an IPv6 address."
    required: false
    default: false
    version_added: "2.5"
  wait:
    description:
      - "When specified,I(state=present) module will wait for subnet to be in available state before continuing."
    required: false
    default: true
    version_added: "2.5"
  wait_timeout:
    description:
      - "Number of seconds to wait for subnet to become available I(wait=True)."
    required: false
    default: 300
    version_added: "2.5"
  purge_tags:
    description:
      - Whether or not to remove tags that do not appear in the I(tags) list. Defaults to true.
    required: false
    default: true
    version_added: "2.5"
extends_documentation_fragment:
    - aws
    - ec2
'''

EXAMPLES = '''
# Note: These examples do not set authentication details, see the AWS Guide for details.

- name: Create subnet for database servers
  ec2_vpc_subnet:
    state: present
    vpc_id: vpc-123456
    cidr: 10.0.1.16/28
    resource_tags:
      Name: Database Subnet
  register: database_subnet

- name: Remove subnet for database servers
  ec2_vpc_subnet:
    state: absent
    vpc_id: vpc-123456
    cidr: 10.0.1.16/28

- name: Create subnet with IPv6 block assigned
  ec2_vpc_subnet:
    state: present
    vpc_id: vpc-123456
    cidr: 10.1.100.0/24
    ipv6_cidr: 2001:db8:0:102::/64

- name: Remove IPv6 block assigned to subnet
  ec2_vpc_subnet:
    state: present
    vpc_id: vpc-123456
    cidr: 10.1.100.0/24
    ipv6_cidr: ''
'''

RETURN = '''
subnet:
    description: Dictionary of subnet values
    returned: I(state=present)
    type: complex
    contains:
        id:
            description: Subnet resource id
            returned: I(state=present)
            type: string
            sample: subnet-b883b2c4
        cidr_block:
            description: The IPv4 CIDR of the Subnet
            returned: I(state=present)
            type: string
            sample: "10.0.0.0/16"
        ipv6_cidr_block:
            description: The IPv6 CIDR block actively associated with the Subnet
            returned: I(state=present)
            type: string
            sample: "2001:db8:0:102::/64"
        availability_zone:
            description: Availability zone of the Subnet
            returned: I(state=present)
            type: string
            sample: us-east-1a
        state:
            description: state of the Subnet
            returned: I(state=present)
            type: string
            sample: available
        tags:
            description: tags attached to the Subnet, includes name
            returned: I(state=present)
            type: dict
            sample: {"Name": "My Subnet", "env": "staging"}
        map_public_ip_on_launch:
            description: whether public IP is auto-assigned to new instances
            returned: I(state=present)
            type: boolean
            sample: false
        assign_ipv6_address_on_creation:
            description: whether IPv6 address is auto-assigned to new instances
            returned: I(state=present)
            type: boolean
            sample: false
        vpc_id:
            description: the id of the VPC where this Subnet exists
            returned: I(state=present)
            type: string
            sample: vpc-67236184
        available_ip_address_count:
            description: number of available IPv4 addresses
            returned: I(state=present)
            type: string
            sample: 251
        default_for_az:
            description: indicates whether this is the default Subnet for this Availability Zone
            returned: I(state=present)
            type: boolean
            sample: false
        ipv6_association_id:
            description: The IPv6 association ID for the currently associated CIDR
            returned: I(state=present)
            type: string
            sample: subnet-cidr-assoc-b85c74d2
        ipv6_cidr_block_association_set:
            description: An array of IPv6 cidr block association set information.
            returned: I(state=present)
            type: complex
            contains:
                association_id:
                    description: The association ID
                    returned: always
                    type: string
                ipv6_cidr_block:
                    description: The IPv6 CIDR block that is associated with the subnet.
                    returned: always
                    type: string
                ipv6_cidr_block_state:
                    description: A hash/dict that contains a single item. The state of the cidr block association.
                    returned: always
                    type: dict
                    contains:
                        state:
                            description: The CIDR block association state.
                            returned: always
                            type: string
'''


import time
import traceback

try:
    import botocore
except ImportError:
    pass  # caught by imported boto3

from ansible.module_utils.aws.core import AnsibleAWSModule
from ansible.module_utils.ec2 import (ansible_dict_to_boto3_filter_list, ansible_dict_to_boto3_tag_list,
                                      ec2_argument_spec, camel_dict_to_snake_dict, get_aws_connection_info,
                                      boto3_conn, boto3_tag_list_to_ansible_dict, compare_aws_tags, AWSRetry)


def get_subnet_info(subnet):
    if 'Subnets' in subnet:
        return [get_subnet_info(s) for s in subnet['Subnets']]
    elif 'Subnet' in subnet:
        subnet = camel_dict_to_snake_dict(subnet['Subnet'])
    else:
        subnet = camel_dict_to_snake_dict(subnet)

    if 'tags' in subnet:
        subnet['tags'] = boto3_tag_list_to_ansible_dict(subnet['tags'])
    else:
        subnet['tags'] = dict()

    if 'subnet_id' in subnet:
        subnet['id'] = subnet['subnet_id']
        del subnet['subnet_id']

    subnet['ipv6_cidr_block'] = ''
    subnet['ipv6_association_id'] = ''
    ipv6set = subnet.get('ipv6_cidr_block_association_set')
    if ipv6set:
        for item in ipv6set:
            if item.get('ipv6_cidr_block_state', {}).get('state') in ('associated', 'associating'):
                subnet['ipv6_cidr_block'] = item['ipv6_cidr_block']
                subnet['ipv6_association_id'] = item['association_id']

    return subnet


@AWSRetry.exponential_backoff()
def describe_subnets_with_backoff(client, **params):
    return client.describe_subnets(**params)


def create_subnet(conn, module, vpc_id, cidr, ipv6_cidr=None, az=None):
    wait = module.params['wait']
    wait_timeout = module.params['wait_timeout']

    params = dict(VpcId=vpc_id,
                  CidrBlock=cidr)

    if ipv6_cidr:
        params['Ipv6CidrBlock'] = ipv6_cidr

    if az:
        params['AvailabilityZone'] = az

    try:
        subnet = get_subnet_info(conn.create_subnet(**params))
    except (botocore.exceptions.ClientError, botocore.exceptions.BotoCoreError) as e:
        module.fail_json_aws(e, msg="Couldn't create subnet")

    # Sometimes AWS takes its time to create a subnet and so using
    # new subnets's id to do things like create tags results in
    # exception.
    if wait and subnet.get('state') != 'available':
        delay = 5
        max_attempts = wait_timeout / delay
        waiter_config = dict(Delay=delay, MaxAttempts=max_attempts)
        waiter = conn.get_waiter('subnet_available')
        try:
            waiter.wait(SubnetIds=[subnet['id']], WaiterConfig=waiter_config)
            subnet['state'] = 'available'
        except (botocore.exceptions.ClientError, botocore.exceptions.BotoCoreError) as e:
            module.fail_json(msg="Create subnet action timed out waiting for Subnet to become available.")

    return subnet


def ensure_tags(conn, module, subnet, tags, purge_tags):
    changed = False

    filters = ansible_dict_to_boto3_filter_list({'resource-id': subnet['id'], 'resource-type': 'subnet'})
    try:
        cur_tags = conn.describe_tags(Filters=filters)
    except (botocore.exceptions.ClientError, botocore.exceptions.BotoCoreError) as e:
        module.fail_json_aws(e, msg="Couldn't describe tags")

    to_update, to_delete = compare_aws_tags(boto3_tag_list_to_ansible_dict(cur_tags.get('Tags')), tags, purge_tags)

    if to_update:
        try:
            if not module.check_mode:
                conn.create_tags(Resources=[subnet['id']], Tags=ansible_dict_to_boto3_tag_list(to_update))

            changed = True
        except (botocore.exceptions.ClientError, botocore.exceptions.BotoCoreError) as e:
            module.fail_json_aws(e, msg="Couldn't create tags")

    if to_delete:
        try:
            if not module.check_mode:
                tags_list = []
                for key in to_delete:
                    tags_list.append({'Key': key})

                conn.delete_tags(Resources=[subnet['id']], Tags=tags_list)

            changed = True
        except (botocore.exceptions.ClientError, botocore.exceptions.BotoCoreError) as e:
            module.fail_json_aws(e, msg="Couldn't delete tags")

    return changed


def ensure_map_public(conn, module, subnet, map_public, check_mode):
    if check_mode:
        return
    try:
        conn.modify_subnet_attribute(SubnetId=subnet['id'], MapPublicIpOnLaunch={'Value': map_public})
    except (botocore.exceptions.ClientError, botocore.exceptions.BotoCoreError) as e:
        module.fail_json_aws(e, msg="Couldn't modify subnet attribute")


def ensure_assign_ipv6_on_create(conn, module, subnet, assign_instances_ipv6, check_mode):
    if check_mode:
        return

    try:
        conn.modify_subnet_attribute(SubnetId=subnet['id'], AssignIpv6AddressOnCreation={'Value': assign_instances_ipv6})
    except (botocore.exceptions.ClientError, botocore.exceptions.BotoCoreError) as e:
        module.fail_json_aws(e, msg="Couldn't modify subnet attribute")


def disassociate_ipv6_cidr(conn, module, subnet):
    if subnet.get('assign_ipv6_address_on_creation'):
        ensure_assign_ipv6_on_create(conn, module, subnet, False, False)

    try:
        conn.disassociate_subnet_cidr_block(AssociationId=subnet['ipv6_association_id'])
    except (botocore.exceptions.ClientError, botocore.exceptions.BotoCoreError) as e:
        module.fail_json_aws(e, msg="Couldn't disassociate ipv6 cidr block id {0} from subnet {1}"
                             .format(subnet['ipv6_association_id'], subnet['id']))


def ensure_ipv6_cidr_block(conn, module, subnet, ipv6_cidr, check_mode):
    changed = False

    if subnet['ipv6_association_id'] and not ipv6_cidr:
        if not check_mode:
            disassociate_ipv6_cidr(conn, module, subnet)
        changed = True

    if ipv6_cidr:
        filters = ansible_dict_to_boto3_filter_list({'ipv6-cidr-block-association.ipv6-cidr-block': ipv6_cidr,
                                                     'vpc-id': subnet['vpc_id']})

        try:
            check_subnets = get_subnet_info(describe_subnets_with_backoff(conn, Filters=filters))
        except (botocore.exceptions.ClientError, botocore.exceptions.BotoCoreError) as e:
            module.fail_json_aws(e, msg="Couldn't get subnet info")

        if check_subnets and check_subnets[0]['ipv6_cidr_block']:
            module.fail_json(msg="The IPv6 CIDR '{0}' conflicts with another subnet".format(ipv6_cidr))

        if subnet['ipv6_association_id']:
            if not check_mode:
                disassociate_ipv6_cidr(conn, module, subnet)
            changed = True

        try:
            if not check_mode:
                associate_resp = conn.associate_subnet_cidr_block(SubnetId=subnet['id'], Ipv6CidrBlock=ipv6_cidr)
            changed = True
        except (botocore.exceptions.ClientError, botocore.exceptions.BotoCoreError) as e:
            module.fail_json_aws(e, msg="Couldn't associate ipv6 cidr {0} to {1}".format(ipv6_cidr, subnet['id']))

        if associate_resp.get('Ipv6CidrBlockAssociation', {}).get('AssociationId'):
            subnet['ipv6_association_id'] = associate_resp['Ipv6CidrBlockAssociation']['AssociationId']
            subnet['ipv6_cidr_block'] = associate_resp['Ipv6CidrBlockAssociation']['Ipv6CidrBlock']
            if subnet['ipv6_cidr_block_association_set']:
                subnet['ipv6_cidr_block_association_set'][0] = camel_dict_to_snake_dict(associate_resp['Ipv6CidrBlockAssociation'])
            else:
                subnet['ipv6_cidr_block_association_set'].append(camel_dict_to_snake_dict(associate_resp['Ipv6CidrBlockAssociation']))

    return changed


def get_matching_subnet(conn, module, vpc_id, cidr):
    filters = ansible_dict_to_boto3_filter_list({'vpc-id': vpc_id, 'cidr-block': cidr})
    try:
        subnets = get_subnet_info(describe_subnets_with_backoff(conn, Filters=filters))
    except (botocore.exceptions.ClientError, botocore.exceptions.BotoCoreError) as e:
        module.fail_json_aws(e, msg="Couldn't get matching subnet")

    if subnets:
        return subnets[0]

    return None


def ensure_subnet_present(conn, module):
    subnet = get_matching_subnet(conn, module, module.params['vpc_id'], module.params['cidr'])
    changed = False
    if subnet is None:
        if not module.check_mode:
            subnet = create_subnet(conn, module, module.params['vpc_id'], module.params['cidr'], ipv6_cidr=module.params['ipv6_cidr'], az=module.params['az'])
        changed = True
        # Subnet will be None when check_mode is true
        if subnet is None:
            return {
                'changed': changed,
                'subnet': {}
            }

    if module.params['ipv6_cidr'] != subnet.get('ipv6_cidr_block'):
        if ensure_ipv6_cidr_block(conn, module, subnet, module.params['ipv6_cidr'], module.check_mode):
            changed = True

    if module.params['map_public'] != subnet['map_public_ip_on_launch']:
        ensure_map_public(conn, module, subnet, module.params['map_public'], module.check_mode)
        changed = True

    if module.params['assign_instances_ipv6'] != subnet.get('assign_ipv6_address_on_creation'):
        ensure_assign_ipv6_on_create(conn, module, subnet, module.params['assign_instances_ipv6'], module.check_mode)
        changed = True

    if module.params['tags'] != subnet['tags']:
        if ensure_tags(conn, module, subnet, module.params['tags'], module.params['purge_tags']):
            changed = True

    subnet = get_matching_subnet(conn, module, module.params['vpc_id'], module.params['cidr'])

    return {
        'changed': changed,
        'subnet': subnet
    }


def ensure_subnet_absent(conn, module):
    subnet = get_matching_subnet(conn, module, module.params['vpc_id'], module.params['cidr'])
    if subnet is None:
        return {'changed': False}

    try:
        if not module.check_mode:
            conn.delete_subnet(SubnetId=subnet['id'])
        return {'changed': True}
    except (botocore.exceptions.ClientError, botocore.exceptions.BotoCoreError) as e:
        module.fail_json_aws(e, msg="Couldn't delete subnet")


def main():
    argument_spec = ec2_argument_spec()
    argument_spec.update(
        dict(
            az=dict(default=None, required=False),
            cidr=dict(default=None, required=True),
            ipv6_cidr=dict(default='', required=False),
            state=dict(default='present', choices=['present', 'absent']),
            tags=dict(default={}, required=False, type='dict', aliases=['resource_tags']),
            vpc_id=dict(default=None, required=True),
            map_public=dict(default=False, required=False, type='bool'),
            assign_instances_ipv6=dict(default=False, required=False, type='bool'),
            wait=dict(type='bool', default=True),
            wait_timeout=dict(type='int', default=300, required=False),
            purge_tags=dict(default=True, type='bool')
        )
    )

    required_if = [('assign_instances_ipv6', True, ['ipv6_cidr'])]

    module = AnsibleAWSModule(argument_spec=argument_spec, supports_check_mode=True, required_if=required_if)

    if module.params.get('assign_instances_ipv6') and not module.params.get('ipv6_cidr'):
        module.fail_json(msg="assign_instances_ipv6 is True but ipv6_cidr is None or an empty string")

    region, ec2_url, aws_connect_params = get_aws_connection_info(module, boto3=True)
    connection = boto3_conn(module, conn_type='client', resource='ec2', region=region, endpoint=ec2_url, **aws_connect_params)

    state = module.params.get('state')

    try:
        if state == 'present':
            result = ensure_subnet_present(connection, module)
        elif state == 'absent':
            result = ensure_subnet_absent(connection, module)
    except botocore.exceptions.ClientError as e:
        module.fail_json(msg=e.message, exception=traceback.format_exc(),
                         **camel_dict_to_snake_dict(e.response))

    module.exit_json(**result)


if __name__ == '__main__':
    main()
