#!/usr/bin/python

# (c) 2017, Vitaliy Zhhuta <zhhuta () gmail.com>
# insipred by Kamil Szczygiel <kamil.szczygiel () intel.com> influxdb_database module
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)

from __future__ import absolute_import, division, print_function

__metaclass__ = type

ANSIBLE_METADATA = {'metadata_version': '1.1',
                    'status': ['preview'],
                    'supported_by': 'community'}

DOCUMENTATION = '''
---
module: influxdb_user
short_description: Manage InfluxDB users
description:
    - Manage InfluxDB users
version_added: 2.5
author: "Vitaliy Zhhuta (@zhhuta)"
requirements:
    - "python >= 2.6"
    - "influxdb >= 0.9"
options:
    hostname:
        description:
            - The hostname or IP address on which InfluxDB server is listening
        default: localhost
        required: false
    user_name:
        description:
            - User that we want to create
        default: None
        required: True
    user_password:
        description:
            - Password that we want to set for user
        default: None
        required: false
    admin:
        description:
            - specify if user should be admin
        default: False
        required: false

    port:
        description:
            - The port on which InfluxDB server is listening
        default: 8086
        required: false
    state:
        description:
            - Determines if the database should be created or destroyed
        choices: ['present', 'absent']
        default: present
        required: false
    username:
        description:
            - user to auth with influxdb
        default: root
        required: false
    password:
        description:
            - password to auth username with influxdb
        default: root
        required: false
'''

EXAMPLES = '''
# Example influxdb_user command from Ansible Playbooks
- name: Create User
  influxdb_user:
      user_name: "{{influxdb_user_name}}"
      user_password: "{{influxdb_user_password}}"
      state: present

- name: Destroy User
  influxdb_user:
      user_name: "{{influxdb_user_name}}"
      username: "{{influxdb_password}}"
      password: "{{influxdb_user_password}}"
      state: absent

- name: Create user on dest host
  influxdb_user:
      hostname: "{{influxdb_ip_address}}"
      username: "{{influxdb_username}}"
      password: "{{influxdb_password}}"
      user_name: "{{influxdb_user_name}}"
      user_password: "{{influxdb_user_password}}"
      admin: true
      state: present
'''

RETURN = '''
#only defaults
'''

import ansible.module_utils.urls
from ansible.module_utils.basic import AnsibleModule
from ansible.module_utils.influxdb import InfluxDb


def find_user(module, client, user_name):
    name = None

    try:
        names = client.get_list_users()
        for u_name in names:
            if u_name['user'] == user_name:
                name = u_name
                break
    except ansible.module_utils.urls.ConnectionError as e:
        module.fail_json(msg=str(e))
    return name


def create_user(module, client, user_name, user_password, admin):
    if not module.check_mode:
        try:
            client.create_user(user_name, user_password, admin)
        except ansible.module_utils.urls.ConnectionError as e:
            module.fail_json(msg=str(e))

    module.exit_json(changed=True)


def drop_user(module, client, user_name):
    if not module.check_mode:
        try:
            client.drop_user(user_name)
        except client.InfluxDBClientError as e:
            module.fail_json(msg=e.content)

    module.exit_json(changed=True)


def main():
    argument_spec = InfluxDb.influxdb_argument_spec()
    argument_spec.update(
        state=dict(default='present', type='str', choices=['present', 'absent']),
        user_name=dict(required=True, type='str'),
        user_password=dict(required=False, type='str', no_log=True),
        admin=dict(default='False', type='bool')

    )
    module = AnsibleModule(
        argument_spec=argument_spec,
        supports_check_mode=True
    )

    state = module.params['state']
    user_name = module.params['user_name']
    user_password = module.params['user_password']
    admin = module.params['admin']
    influxdb = InfluxDb(module)
    client = influxdb.connect_to_influxdb()
    user = find_user(module, client, user_name)

    if state == 'present':
        if user:
            module.exit_json(changed=False)
        else:
            create_user(module, client, user_name, user_password, admin)

    if state == 'absent':
        if user:
            drop_user(module, client, user_name)
        else:
            module.exit_json(changed=False)


if __name__ == '__main__':
    main()
