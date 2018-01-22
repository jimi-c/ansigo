#!/usr/bin/python
# -*- coding: utf-8 -*-

# This file is part of Ansible

# Copyright (c) 2017 Ansible Project
# GNU General Public License v3.0+ (see COPYING or https://www.gnu.org/licenses/gpl-3.0.txt)


ANSIBLE_METADATA = {'metadata_version': '1.1',
                    'status': ['preview'],
                    'supported_by': 'community'}


DOCUMENTATION = r'''
---
module: win_certificate_store
version_added: '2.5'
short_description: Manages the certificate store
description:
- Used to import/export and remove certificates and keys from the local
  certificate store.
- This module is not used to create certificates and will only manage existing
  certs as a file or in the store.
- It can be used to import PEM, DER, P7B, PKCS12 (PFX) certificates and export
  PEM, DER and PKCS12 certificates.
options:
  state:
    description:
    - If C(present), will ensure that the certificate at I(path) is imported
      into the certificate store specified.
    - If C(absent), will ensure that the certificate specified by I(thumbprint)
      or the thumbprint of the cert at I(path) is removed from the store
      specified.
    - If C(exported), will ensure the file at I(path) is a certificate
      specified by I(thumbprint).
    - When exporting a certificate, if I(path) is a directory then the module
      will fail, otherwise the file will be replaced if needed.
    default: present
    choices:
    - present
    - absent
    - exported
  path:
    description:
    - The path to a certificate file.
    - This is required when I(state) is C(present) or C(exported).
    - When I(state) is C(absent) and I(thumbprint) is not specified, the
      thumbprint is derived from the certificate at this path.
  thumbprint:
    description:
    - The thumbprint as a hex string to either export or remove.
    - See the examples for how to specify the thumbprint.
  store_name:
    description:
    - The store name to use when importing a certificate or searching for a
      certificate.
    default: My
    choices:
    - AddressBook
    - AuthRoot
    - CertificateAuthority
    - Disallowed
    - My
    - Root
    - TrustedPeople
    - TrustedPublisher
  store_location:
    description:
    - The store location to use when importing a certificate or searching for a
      certificate.
    default: LocalMachine
    choices:
    - CurrentUser
    - LocalMachine
  password:
    description:
    - The password of the pkcs12 certificate key.
    - This is used when reading a pkcs12 certificate file or the password to
      set when C(state=exported) and C(file_type=pkcs12).
    - If the pkcs12 file has no password set or no password should be set on
      the exported file, do not set this option.
  key_exportable:
    description:
    - Whether to allow the private key to be exported.
    - If C(no), then this module and other process will only be able to export
      the certificate and the private key cannot be exported.
    - Used when C(state=present) only.
    type: bool
    default: 'yes'
  key_storage:
    description:
    - Specifies where Windows will store the private key when it is imported.
    - When set to C(default), the default option as set by Windows is used.
    - When set to C(machine), the key is stored in a path accessible by various
      users.
    - When set to C(user), the key is stored in a path only accessible by the
      current user.
    - Used when C(state=present) only and cannot be changed once imported.
    - See U(https://msdn.microsoft.com/en-us/library/system.security.cryptography.x509certificates.x509keystorageflags.aspx)
      for more details.
    choices:
    - default
    - machine
    - user
    default: default
  file_type:
    description:
    - The file type to export the certificate as when C(state=exported).
    - C(der) is a binary ASN.1 encoded file.
    - C(pem) is a base64 encoded file of a der file in the OpenSSL form.
    - C(pkcs12) (also known as pfx) is a binary container that contains both
      the certificate and private key unlike the other options.
    - When C(pkcs12) is set and the private key is not exportable or accessible
      by the current user, it will throw an exception.
    choices:
    - der
    - pem
    - pkcs12
    default: der
notes:
- Some actions on PKCS12 certificates and keys may fail with the error
  C(the specified network password is not correct), either use CredSSP or
  Kerberos with credential delegation, or use C(become) to bypass these
  restrictions.
- The certificates must be located on the Windows host to be set with I(path).
author:
- Jordan Borean (@jborean93)
'''

EXAMPLES = r'''
- name: import a certificate
  win_certificate_store:
    path: C:\temp\cert.pem
    state: present

- name: import pfx certificate that is password protected
  win_certificate_store:
    path: C:\temp\cert.pfx
    state: present
    password: VeryStrongPasswordHere!
  become: yes
  become_method: runas

- name: import pfx certificate without password and set private key as un-exportable
  win_certificate_store:
    path: C:\temp\cert.pfx
    state: present
    key_exportable: no
  # usually you don't set this here but it is for illustrative purposes
  vars:
    ansible_winrm_transport: credssp

- name: remove a certificate based on file thumbprint
  win_certificate_store:
    path: C:\temp\cert.pem
    state: absent

- name: remove a certificate based on thumbprint
  win_certificate_store:
    thumbprint: BD7AF104CF1872BDB518D95C9534EA941665FD27
    state: absent

- name: remove certificate based on thumbprint is CurrentUser/TrustedPublishers store
  win_certificate_store:
    thumbprint: BD7AF104CF1872BDB518D95C9534EA941665FD27
    state: absent
    store_location: CurrentUser
    store_name: TrustedPublisher

- name: export certificate as der encoded file
  win_certificate_store:
    path: C:\temp\cert.cer
    state: exported
    file_type: der

- name: export certificate and key as pfx encoded file
  win_certificate_store:
    path: C:\temp\cert.pfx
    state: exported
    file_type: pkcs12
    password: AnotherStrongPass!
  become: yes
  become_method: runas
  become_user: SYSTEM
'''

RETURN = r'''
thumbprints:
  description: A list of certificate thumbprints that were touched by the
    module.
  returned: success
  type: list
  sample: ["BC05633694E675449136679A658281F17A191087"]
'''
