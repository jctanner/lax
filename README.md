![image](https://github.com/jctanner/lax/assets/1869705/a8d0c3ca-22f8-470a-b50e-79d7a9ed58e9)

LAX is a sleek and efficient CLI tool designed to replace Ansible-Galaxy. It simplifies and accelerates your workflow, making it easier to manage your Ansible roles and collections.

## Features

- **Simplicity**: Easy to host and to install collections&roles
- **Speed**: Aims to be faster in every way possible
- **Efficiency**: No need for a heavy CRUD app to host content

## Benchmarks

Lax is faster than ansible-galaxy ...

```
# lax collection install --server=https://tannerjc.net/galaxy sivel.acd
real    0m21.897s
user    0m4.660s
sys     0m2.089s


# ansible-galaxy collection install sivel.acd
real    4m35.545s
user    1m12.326s
sys     0m8.316s
```

Why is it faster? It's not because it's written in golang. It's because ansible-galaxy and the galaxy server api have been designed to make a LOT of network requests to find information about collections and to build dependency maps ...

```
$ rm -rf ~/.ansible; ansible-galaxy collection install -vvvv sivel.acd
...
Calling Galaxy at https://galaxy.ansible.com/api/
Found API version 'v3, pulp-v3, v1' with Galaxy server default (https://galaxy.ansible.com/api/)
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/sivel/acd/
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/sivel/acd/versions/?limit=100
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/sivel/acd/versions/6.5.0/
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/awx/awx/
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/awx/awx/versions/?limit=100
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/frr/frr/
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/frr/frr/versions/?limit=100
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/cisco/aci/
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/cisco/aci/versions/?limit=100
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/cisco/asa/
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/cisco/asa/versions/?limit=100
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/cisco/ios/
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/cisco/ios/versions/?limit=100
Calling Galaxy at https://galaxy.ansible.com/api/v3/plugin/ansible/content/published/collections/index/cisco/ios/versions/?limit=100&offset=100
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/cisco/ise/
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/cisco/ise/versions/?limit=100
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/cisco/mso/
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/cisco/mso/versions/?limit=100
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/cisco/nso/
Calling Galaxy at https://galaxy.ansible.com/api/v3/collections/cisco/nso/versions/?limit=100
...
```

ansible-galaxy has to fetch at least 2 api endpoints for every collection it needs to install. To mitigate this requirement, the client tries to cache api data so that subsequent installs aren't so slow. However, the API design still remains a bottleneck.

Lax on the other hand is designed to work more like yum/dnf in that the "server" isn't an API, it is instead just an http based file share. Lax checks for a metadata.json file at the root of the file share and from that definition, fetches a tar.gz of all collection metadata (including dependencies). Using just that metadata, lax compiles a depedency tree and then fetches all the necessary tarballs from the fileshare.

## Installation

See the [User Guide](docs/USERGUIDE.md)

## Usage

See the [User Guide](docs/USERGUIDE.md)

## Contributing

TBD

## License

TBD

## Contact

TBD
