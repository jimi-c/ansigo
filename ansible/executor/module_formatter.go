package executor

import (
  "archive/zip"
  "bytes"
  "encoding/base64"
  "encoding/json"
  "io"
  "io/ioutil"
  "os"
  "path"
  "path/filepath"
  "regexp"
  "strings"
  "../playbook"
  "../plugins"
)

const ANSIBALLZ_TEMPLATE = `%{shebang}s
%{encoding}s
ANSIBALLZ_WRAPPER = True # For test-module script to tell this is a ANSIBALLZ_WRAPPER
# This code is part of Ansible, but is an independent component.
# The code in this particular templatable string, and this templatable string
# only, is BSD licensed.  Modules which end up using this snippet, which is
# dynamically combined together by Ansible still belong to the author of the
# module, and they may assign their own license to the complete work.
#
# Copyright (c), James Cammarata, 2016
# Copyright (c), Toshio Kuratomi, 2016
#
# Redistribution and use in source and binary forms, with or without modification,
# are permitted provided that the following conditions are met:
#
#    * Redistributions of source code must retain the above copyright
#      notice, this list of conditions and the following disclaimer.
#    * Redistributions in binary form must reproduce the above copyright notice,
#      this list of conditions and the following disclaimer in the documentation
#      and/or other materials provided with the distribution.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
# ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
# WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
# IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT,
# INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
# PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
# INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
# LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE
# USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
import os
import os.path
import sys
import __main__

# For some distros and python versions we pick up this script in the temporary
# directory.  This leads to problems when the ansible module masks a python
# library that another import needs.  We have not figured out what about the
# specific distros and python versions causes this to behave differently.
#
# Tested distros:
# Fedora23 with python3.4  Works
# Ubuntu15.10 with python2.7  Works
# Ubuntu15.10 with python3.4  Fails without this
# Ubuntu16.04.1 with python3.5  Fails without this
# To test on another platform:
# * use the copy module (since this shadows the stdlib copy module)
# * Turn off pipelining
# * Make sure that the destination file does not exist
# * ansible ubuntu16-test -m copy -a 'src=/etc/motd dest=/var/tmp/m'
# This will traceback in shutil.  Looking at the complete traceback will show
# that shutil is importing copy which finds the ansible module instead of the
# stdlib module
scriptdir = None
try:
    scriptdir = os.path.dirname(os.path.realpath(__main__.__file__))
except (AttributeError, OSError):
    # Some platforms don't set __file__ when reading from stdin
    # OSX raises OSError if using abspath() in a directory we don't have
    # permission to read (realpath calls abspath)
    pass
if scriptdir is not None:
    sys.path = [p for p in sys.path if p != scriptdir]

import base64
import shutil
import zipfile
import tempfile
import subprocess

if sys.version_info < (3,):
    bytes = str
    PY3 = False
else:
    unicode = str
    PY3 = True
try:
    # Python-2.6+
    from io import BytesIO as IOStream
except ImportError:
    # Python < 2.6
    from StringIO import StringIO as IOStream

ZIPDATA = """%{zipped_data}s"""

def invoke_module(module, modlib_path, json_params):
    pythonpath = os.environ.get('PYTHONPATH')
    if pythonpath:
        os.environ['PYTHONPATH'] = ':'.join((modlib_path, pythonpath))
    else:
        os.environ['PYTHONPATH'] = modlib_path

    p = subprocess.Popen(["%{interpreter}s", module], env=os.environ, shell=False, stdout=subprocess.PIPE, stderr=subprocess.PIPE, stdin=subprocess.PIPE)
    (stdout, stderr) = p.communicate(json_params)

    if not isinstance(stderr, (bytes, unicode)):
        stderr = stderr.read()
    if not isinstance(stdout, (bytes, unicode)):
        stdout = stdout.read()
    if PY3:
        sys.stderr.buffer.write(stderr)
        sys.stdout.buffer.write(stdout)
    else:
        sys.stderr.write(stderr)
        sys.stdout.write(stdout)
    return p.returncode

def debug(command, zipped_mod, json_params):
    # The code here normally doesn't run.  It's only used for debugging on the
    # remote machine.
    #
    # The subcommands in this function make it easier to debug ansiballz
    # modules.  Here's the basic steps:
    #
    # Run ansible with the environment variable: ANSIBLE_KEEP_REMOTE_FILES=1 and -vvv
    # to save the module file remotely::
    #   $ ANSIBLE_KEEP_REMOTE_FILES=1 ansible host1 -m ping -a 'data=october' -vvv
    #
    # Part of the verbose output will tell you where on the remote machine the
    # module was written to::
    #   [...]
    #   <host1> SSH: EXEC ssh -C -q -o ControlMaster=auto -o ControlPersist=60s -o KbdInteractiveAuthentication=no -o
    #   PreferredAuthentications=gssapi-with-mic,gssapi-keyex,hostbased,publickey -o PasswordAuthentication=no -o ConnectTimeout=10 -o
    #   ControlPath=/home/badger/.ansible/cp/ansible-ssh-%h-%p-%r -tt rhel7 '/bin/sh -c '"'"'LANG=en_US.UTF-8 LC_ALL=en_US.UTF-8
    #   LC_MESSAGES=en_US.UTF-8 /usr/bin/python /home/badger/.ansible/tmp/ansible-tmp-1461173013.93-9076457629738/ping'"'"''
    #   [...]
    #
    # Login to the remote machine and run the module file via from the previous
    # step with the explode subcommand to extract the module payload into
    # source files::
    #   $ ssh host1
    #   $ /usr/bin/python /home/badger/.ansible/tmp/ansible-tmp-1461173013.93-9076457629738/ping explode
    #   Module expanded into:
    #   /home/badger/.ansible/tmp/ansible-tmp-1461173408.08-279692652635227/ansible
    #
    # You can now edit the source files to instrument the code or experiment with
    # different parameter values.  When you're ready to run the code you've modified
    # (instead of the code from the actual zipped module), use the execute subcommand like this::
    #   $ /usr/bin/python /home/badger/.ansible/tmp/ansible-tmp-1461173013.93-9076457629738/ping execute

    # Okay to use __file__ here because we're running from a kept file
    basedir = os.path.join(os.path.abspath(os.path.dirname(__file__)), 'debug_dir')
    args_path = os.path.join(basedir, 'args')
    script_path = os.path.join(basedir, 'ansible_module_%{module_name}s.py')

    if command == 'explode':
        # transform the ZIPDATA into an exploded directory of code and then
        # print the path to the code.  This is an easy way for people to look
        # at the code on the remote machine for debugging it in that
        # environment
        z = zipfile.ZipFile(zipped_mod)
        for filename in z.namelist():
            if filename.startswith('/'):
                raise Exception('Something wrong with this module zip file: should not contain absolute paths')

            dest_filename = os.path.join(basedir, filename)
            if dest_filename.endswith(os.path.sep) and not os.path.exists(dest_filename):
                os.makedirs(dest_filename)
            else:
                directory = os.path.dirname(dest_filename)
                if not os.path.exists(directory):
                    os.makedirs(directory)
                f = open(dest_filename, 'wb')
                f.write(z.read(filename))
                f.close()

        # write the args file
        f = open(args_path, 'wb')
        f.write(json_params)
        f.close()

        print('Module expanded into:')
        print('%s' % basedir)
        exitcode = 0

    elif command == 'execute':
        # Execute the exploded code instead of executing the module from the
        # embedded ZIPDATA.  This allows people to easily run their modified
        # code on the remote machine to see how changes will affect it.
        # This differs slightly from default Ansible execution of Python modules
        # as it passes the arguments to the module via a file instead of stdin.

        # Set pythonpath to the debug dir
        pythonpath = os.environ.get('PYTHONPATH')
        if pythonpath:
            os.environ['PYTHONPATH'] = ':'.join((basedir, pythonpath))
        else:
            os.environ['PYTHONPATH'] = basedir

        p = subprocess.Popen(["%{interpreter}s", script_path, args_path],
                env=os.environ, shell=False, stdout=subprocess.PIPE,
                stderr=subprocess.PIPE, stdin=subprocess.PIPE)
        (stdout, stderr) = p.communicate()

        if not isinstance(stderr, (bytes, unicode)):
            stderr = stderr.read()
        if not isinstance(stdout, (bytes, unicode)):
            stdout = stdout.read()
        if PY3:
            sys.stderr.buffer.write(stderr)
            sys.stdout.buffer.write(stdout)
        else:
            sys.stderr.write(stderr)
            sys.stdout.write(stdout)
        return p.returncode

    elif command == 'excommunicate':
        # This attempts to run the module in-process (by importing a main
        # function and then calling it).  It is not the way ansible generally
        # invokes the module so it won't work in every case.  It is here to
        # aid certain debuggers which work better when the code doesn't change
        # from one process to another but there may be problems that occur
        # when using this that are only artifacts of how we're invoking here,
        # not actual bugs (as they don't affect the real way that we invoke
        # ansible modules)

        # stub the args and python path
        sys.argv = ['%{module_name}s', args_path]
        sys.path.insert(0, basedir)

        from ansible_module_%{module_name}s import main
        main()
        print('WARNING: Module returned to wrapper instead of exiting')
        sys.exit(1)
    else:
        print('WARNING: Unknown debug command.  Doing nothing.')
        exitcode = 0

    return exitcode

if __name__ == '__main__':
    #
    # See comments in the debug() method for information on debugging
    #

    ANSIBALLZ_PARAMS = '''%{params}s'''
    if PY3:
        ANSIBALLZ_PARAMS = ANSIBALLZ_PARAMS.encode('utf-8')
    try:
        # There's a race condition with the controller removing the
        # remote_tmpdir and this module executing under async.  So we cannot
        # store this in remote_tmpdir (use system tempdir instead)
        temp_path = tempfile.mkdtemp(prefix='ansible_')

        zipped_mod = os.path.join(temp_path, 'ansible_modlib.zip')
        modlib = open(zipped_mod, 'wb')
        modlib.write(base64.b64decode(ZIPDATA))
        modlib.close()

        if len(sys.argv) == 2:
            exitcode = debug(sys.argv[1], zipped_mod, ANSIBALLZ_PARAMS)
        else:
            z = zipfile.ZipFile(zipped_mod, mode='r')
            module = os.path.join(temp_path, 'ansible_module_%{module_name}s.py')
            f = open(module, 'wb')
            f.write(z.read('ansible_module_%{module_name}s.py'))
            f.close()

            # When installed via setuptools (including python setup.py install),
            # ansible may be installed with an easy-install.pth file.  That file
            # may load the system-wide install of ansible rather than the one in
            # the module.  sitecustomize is the only way to override that setting.
            z = zipfile.ZipFile(zipped_mod, mode='a')

            # py3: zipped_mod will be text, py2: it's bytes.  Need bytes at the end
            #sitecustomize = u'import sys\\nsys.path.insert(0,"%s")\\n' % zipped_mod
            #sitecustomize = sitecustomize.encode('utf-8')
            # Use a ZipInfo to work around zipfile limitation on hosts with
            # clocks set to a pre-1980 year (for instance, Raspberry Pi)
            #zinfo = zipfile.ZipInfo()
            #zinfo.filename = 'sitecustomize.py'
            #zinfo.date_time = ( %{year}s, %{month}s, %{day}s, %{hour}s, %{minute}s, %{second}s)
            #z.writestr(zinfo, sitecustomize)
            #z.close()

            exitcode = invoke_module(module, zipped_mod, ANSIBALLZ_PARAMS)
    finally:
        try:
            shutil.rmtree(temp_path)
        except (NameError, OSError):
            # tempdir creation probably failed
            pass
    sys.exit(exitcode)
`
var CompiledModuleCache map[string]string = make(map[string]string)
var re = regexp.MustCompile("ansible\\.module_utils\\.([\\w.]+)+")

func CompileDependencies(data string, dependencies []string, archive *zip.Writer) []string {
  exPath := plugins.GetExecutableDir()
  matches := re.FindAllString(data, -1)
  replacer := strings.NewReplacer("ansible.", "", ".", string(os.PathSeparator))
  for _, match := range matches {
    // FIXME: this needs to be read from a well-known module path
    //        instead of relative to the binary executable
    target_path := path.Join(exPath, "..", "modules", replacer.Replace(match))
    stat, err := os.Lstat(target_path)
    // if the path is a directory, use CompileDependencies recursively
    if err == nil && stat.IsDir() {
      // walk the dir and all of the files to the zip
      if playbook.StringPos(target_path, dependencies) != -1 {
        continue
      }
      filepath.Walk(target_path, func(path string, info os.FileInfo, err error) error {
        if err != nil {
          return err
        }
        header, err := zip.FileInfoHeader(info)
        if err != nil {
          return err
        }
        parts := strings.Split(match, ".")
        header.Name = filepath.Join("ansible", "module_utils", parts[len(parts)-1], strings.TrimPrefix(path, target_path))
        if info.IsDir() {
          header.Name += "/"
        } else {
          header.Method = zip.Deflate
        }
        if playbook.StringPos(header.Name, dependencies) != -1 {
          return nil
        }
        dependencies = append(dependencies, header.Name)
        writer, err := archive.CreateHeader(header)
        if err != nil {
          return err
        }
        if info.IsDir() {
          // FIXME: recurse into sub directories?
          return nil
        }
        file, err := os.Open(path)
        if err != nil {
          return err
        }
        defer file.Close()
        _, err = io.Copy(writer, file)
        return err
      })
    } else {
      // otherwise, if the path exists with a .py extension, add it to the zip
      py_path := target_path + ".py"
      if _, err := os.Lstat(py_path); err != nil {
        parts := strings.Split(match, ".")
        py_path = path.Join(exPath, "..", "modules", path.Join(parts[1:len(parts)-1]...) + ".py")
        if _, err := os.Lstat(py_path); err != nil {
          // FIXME: error handling
          continue
        }
        match = strings.Join(parts[1:len(parts)-1], ".")
      }
      dep_data, err := ioutil.ReadFile(py_path)
      if err != nil {
        // FIXME: error handling
      }
      // split original match on .'s and remove the last one
      replacer := strings.NewReplacer(".", string(os.PathSeparator))
      archive_path := replacer.Replace(match) + ".py"
      if playbook.StringPos(archive_path, dependencies) != -1 {
        continue
      }
      dependencies = append(dependencies, archive_path)
      writer, err := archive.Create(archive_path)
      if err != nil {
        // FIXME: handle error
      }
      io.Copy(writer, bytes.NewReader(dep_data))
      dependencies = CompileDependencies(string(dep_data), dependencies, archive)
    }
  }
  return dependencies
}

func CompileModule(name string, args map[string]interface{}) string {
  module_info, ok := playbook.ModuleCache[name]
  if !ok {
    panic("COULDN'T FIND THE MODULE: '" + name + "'")
  }

  zipped_data := ""
  if compiled_module, ok := CompiledModuleCache[name]; ok {
    zipped_data = compiled_module
  } else {
    out_buffer := bytes.NewBufferString("")
    archive := zip.NewWriter(out_buffer)

    data, err := ioutil.ReadFile(module_info.Path)
    if err != nil {
      // FIXME: handle error
    }

    if writer, err := archive.Create("ansible_module_" + name + ".py"); err != nil {
      // FIXME: handle error
    } else {
      io.Copy(writer, bytes.NewReader(data))
    }
    // create base init files
    for _, f_name := range []string{"ansible/__init__.py", "ansible/module_utils/__init__.py"} {
      if writer, err := archive.Create(f_name); err != nil {
        // FIXME: handle error
      } else {
        io.Copy(writer, bytes.NewReader([]byte{}))
      }
    }

    dependencies := make([]string, 0)
    dependencies = CompileDependencies(string(data), dependencies, archive)

    // Add inits for any directories created while archiving
    // dependencies, but for which were not already included
    inits := make([]string, 0)
    for _, dep := range dependencies {
      for dep_dir := path.Dir(dep); dep_dir != "."; dep_dir = path.Dir(dep_dir) {
        init := path.Join(dep_dir, "__init__.py")
        if playbook.StringPos(init, dependencies) != -1 || playbook.StringPos(init, inits) != -1 {
          continue
        } else {
          if writer, err := archive.Create(init); err != nil {
            // FIXME: handle error
          } else {
            io.Copy(writer, bytes.NewReader([]byte{}))
          }
          inits = append(inits, init)
        }
      }
    }
    archive.Close()

    zipped_data = base64.StdEncoding.EncodeToString(out_buffer.Bytes())
    CompiledModuleCache[name] = zipped_data
  }

  var params = map[string]interface{} {
    "ANSIBLE_MODULE_ARGS": args,
  }
  encoded_params, err := json.Marshal(params)
  if err != nil {
    // FIXME: handle error
  }
  var formatting_params = map[string]interface{} {
    "module_name": name,
    "shebang": "#!/usr/bin/python",
    "interpreter": "/usr/bin/python",
    "encoding": "# -*- coding: utf-8 -*-",
    "zipped_data": zipped_data,
    "params": string(encoded_params),
    "year": "2018",
    "month": "1",
    "day": "1",
    "hour": "0",
    "minute": "0",
    "second": "0",
  }
  formatted_string := Tprintf(ANSIBALLZ_TEMPLATE, formatting_params)
  //ioutil.WriteFile("/tmp/module_" + name + ".py", []byte(formatted_string), 0644)
  return formatted_string
}
