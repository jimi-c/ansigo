#!/usr/bin/python
# Copyright (c) 2016 Hewlett-Packard Enterprise Corporation
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function
__metaclass__ = type


ANSIBLE_METADATA = {'metadata_version': '1.1',
                    'status': ['preview'],
                    'supported_by': 'community'}


DOCUMENTATION = '''
---
module: os_user_facts
short_description: Retrieve facts about one or more OpenStack users
extends_documentation_fragment: openstack
version_added: "2.1"
author: "Ricardo Carrillo Cruz (@rcarrillocruz)"
description:
    - Retrieve facts about a one or more OpenStack users
requirements:
    - "python >= 2.6"
    - "shade"
options:
   name:
     description:
        - Name or ID of the user
     required: true
   domain:
     description:
        - Name or ID of the domain containing the user if the cloud supports domains
     required: false
     default: None
   filters:
     description:
        - A dictionary of meta data to use for further filtering.  Elements of
          this dictionary may be additional dictionaries.
     required: false
     default: None
   availability_zone:
     description:
       - Ignored. Present for backwards compatibility
     required: false
'''

EXAMPLES = '''
# Gather facts about previously created users
- os_user_facts:
    cloud: awesomecloud
- debug:
    var: openstack_users

# Gather facts about a previously created user by name
- os_user_facts:
    cloud: awesomecloud
    name: demouser
- debug:
    var: openstack_users

# Gather facts about a previously created user in a specific domain
- os_user_facts:
    cloud: awesomecloud
    name: demouser
    domain: admindomain
- debug:
    var: openstack_users

# Gather facts about a previously created user in a specific domain with filter
- os_user_facts:
    cloud: awesomecloud
    name: demouser
    domain: admindomain
    filters:
      enabled: False
- debug:
    var: openstack_users
'''


RETURN = '''
openstack_users:
    description: has all the OpenStack facts about users
    returned: always, but can be null
    type: complex
    contains:
        id:
            description: Unique UUID.
            returned: success
            type: string
        name:
            description: Name given to the user.
            returned: success
            type: string
        enabled:
            description: Flag to indicate if the user is enabled
            returned: success
            type: bool
        domain_id:
            description: Domain ID containing the user
            returned: success
            type: string
        default_project_id:
            description: Default project ID of the user
            returned: success
            type: string
        email:
            description: Email of the user
            returned: success
            type: string
        username:
            description: Username of the user
            returned: success
            type: string
'''

try:
    import shade
    HAS_SHADE = True
except ImportError:
    HAS_SHADE = False

from ansible.module_utils.basic import AnsibleModule
from ansible.module_utils.openstack import openstack_full_argument_spec


def main():

    argument_spec = openstack_full_argument_spec(
        name=dict(required=False, default=None),
        domain=dict(required=False, default=None),
        filters=dict(required=False, type='dict', default=None),
    )

    module = AnsibleModule(argument_spec)

    if not HAS_SHADE:
        module.fail_json(msg='shade is required for this module')

    try:
        name = module.params['name']
        domain = module.params['domain']
        filters = module.params['filters']

        opcloud = shade.operator_cloud(**module.params)

        if domain:
            try:
                # We assume admin is passing domain id
                dom = opcloud.get_domain(domain)['id']
                domain = dom
            except:
                # If we fail, maybe admin is passing a domain name.
                # Note that domains have unique names, just like id.
                dom = opcloud.search_domains(filters={'name': domain})
                if dom:
                    domain = dom[0]['id']
                else:
                    module.fail_json(msg='Domain name or ID does not exist')

            if not filters:
                filters = {}

            filters['domain_id'] = domain

        users = opcloud.search_users(name, filters)
        module.exit_json(changed=False, ansible_facts=dict(
            openstack_users=users))

    except shade.OpenStackCloudException as e:
        module.fail_json(msg=str(e))


if __name__ == '__main__':
    main()
