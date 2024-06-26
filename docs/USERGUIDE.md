# User guide

## Preface

LAX's primary workflow is to install .tar.gz files from a http or file location. There is not yet a central host for lax, so it is the user's responsibility currently to make their own repo to host the content. However, making a repo is far easier than it sounds.

A LAX repo, aka "server", is a directory with two subdirectories (roles & collections). In the root directory there is a json file named `repometa.json`. LAX inspects the repometa.json to figure out what other index files exist for the repo (usually role_manfiests.tar.gz & collection_manifests.tar.gz). The index files provide LAX with all the information it needs about the available content and their dependencies.

When the user installs a package from the LAX repo, it finds the necessary package and dependencies from the index files then downloads the relevant tar.gz files and extracts them to the appropriate location for `ansible` or `ansible-playbook` to use.

## Installing

The quickest way to install lax right now is with `go install` ...

```
(venv) jtanner@p15:/tmp/test$ docker run -it golang:latest bash
root@a47952ea7696:/go# go install github.com/jctanner/lax/cmd/lax@latest
...
root@a47952ea7696:/go# lax --help
Usage:
  cli [command]

Available Commands:
  collection  Manage collections
  completion  Generate the autocompletion script for the specified shell
  createrepo  Create repository metadata from a directory of artifacts
  galaxy-sync Sync content from galaxy into a lax repo directory
  help        Help about any command
  role        Manage roles

Flags:
  -h, --help   help for cli

Use "cli [command] --help" for more information about a command.
```

Binary distributions will eventually be released on github as the project gets more stable.

## Syncing Content

LAX needs content to install and you can make your own and put them in to a folder under `<folder>/roles` or `<folder>/collections` or you can use the `lax galaxy-sync` subcommand to fetch content from galaxy.ansible.com and github.com.

There's a couple different command options for `lax galaxy-sync` that make the syncing work to your specific needs ...

```
root@a47952ea7696:/go# lax galaxy-sync --help
Sync content from galaxy into a lax repo directory

Usage:
  cli galaxy-sync [flags]

Flags:
      --artifacts             just sync the artifacts
      --collections           just sync collections
      --concurrency int       concurrency (default 1)
      --dest string           where to store the data
  -h, --help                  help for galaxy-sync
      --latest                get only the latest version
      --name string           name
      --namespace string      namespace
  -r, --requirements string   requirements file
      --roles                 just sync roles
      --server string         remote server (default "https://galaxy.ansible.com")

```

If you wanted to just get the `geerlingguy.docker` role, you could simply run ...

```
root@a47952ea7696:/go# lax galaxy-sync --dest=/tmp/foo --roles --namespace=geerlingguy --name=docker --latest
...

root@a47952ea7696:/go# find /tmp/foo
/tmp/foo
/tmp/foo/.cache
/tmp/foo/.cache/roles
/tmp/foo/.cache/roles/2056ecf0.json
/tmp/foo/.cache/roles/4ae72842.json
/tmp/foo/.cache/roles/95904a2e.json
/tmp/foo/.cache/collections
/tmp/foo/collections
/tmp/foo/roles
/tmp/foo/roles/geerlingguy-docker-7.2.0.tar.gz

```

LAX found the geerlingguy.docker role on galaxy.ansible.com and then fetched the "latest" version's .tar.gz from github.com. If the `--latest` flag had not been specified, it would have fetched ALL of the available versions for the role.

The same process can be followed for collections ...

```
root@a47952ea7696:/go# lax galaxy-sync --dest=/tmp/foo --collections --namespace=geerlingguy --name=mac --latest
...
root@a47952ea7696:/go# find /tmp/foo
/tmp/foo
/tmp/foo/.cache
/tmp/foo/.cache/roles
/tmp/foo/.cache/roles/98d7187f.json
/tmp/foo/.cache/roles/a4984ad.json
/tmp/foo/.cache/roles/3ed6394e.json
/tmp/foo/.cache/collections
/tmp/foo/collections
/tmp/foo/collections/geerlingguy-mac-4.0.1.tar.gz
/tmp/foo/roles
```

## Creating a Repo

Assuming you have a directory of content where you've stored .tar.gz files for roles and or collections like so ...

```
root@a47952ea7696:/go# find /tmp/foo -type f
/tmp/foo/.cache/roles/98d7187f.json
/tmp/foo/.cache/roles/a4984ad.json
/tmp/foo/.cache/roles/3ed6394e.json
/tmp/foo/.cache/roles/2056ecf0.json
/tmp/foo/.cache/roles/4ae72842.json
/tmp/foo/.cache/roles/95904a2e.json
/tmp/foo/collections/geerlingguy-mac-4.0.1.tar.gz
/tmp/foo/roles/geerlingguy-docker-7.2.0.tar.gz
```

You can then create a "repository", which really just means compile some index files that contain the metadata about the content in the directory...

```
root@a47952ea7696:/go# lax createrepo --dest=/tmp/foo
...
root@a47952ea7696:/go# find /tmp/foo
/tmp/foo
/tmp/foo/.cache
/tmp/foo/.cache/roles
/tmp/foo/.cache/roles/98d7187f.json
/tmp/foo/.cache/roles/a4984ad.json
/tmp/foo/.cache/roles/3ed6394e.json
/tmp/foo/.cache/roles/2056ecf0.json
/tmp/foo/.cache/roles/4ae72842.json
/tmp/foo/.cache/roles/95904a2e.json
/tmp/foo/.cache/collections
/tmp/foo/collections
/tmp/foo/collections/geerlingguy-mac-4.0.1.tar.gz
/tmp/foo/roles
/tmp/foo/roles/geerlingguy-docker-7.2.0.tar.gz
/tmp/foo/metadata
/tmp/foo/collection_manifests.tar.gz    <-- the MANIFEST.json content for every collection
/tmp/foo/collection_files.tar.gz        <-- a listing of all files in every collection
/tmp/foo/role_manifests.tar.gz          <-- the meta/main.yml content for every role
/tmp/foo/role_files.tar.gz              <-- a listing of all files in every role
/tmp/foo/repometa.json                  <-- maps out the other meta files and includes timestamps
```

This repository is now ready to serve!!!

## Hosting a Repo

LAX aims to be flexible, so the repository directory can live locally OR it can live on an http fileshare you've hosted on the network. There is no special magic to hosting files on the internet and most webserver implemenations can serve out the files. Use rsync or ftp or whatever protocol to send the repository directory to your web host.

## Installing Content From a Repo

LAX's true purpose is to install roles and collections so that ansible and ansible-playbook are able to use them. It will also try to remain fully compatible with the expectations of the `ansible-galaxy` cli for ondisk files.

To install a role from a local repository directory ...

```
root@a47952ea7696:/go# lax role install --server=/tmp/foo geerlingguy.docker
...
root@a47952ea7696:/go# ansible-galaxy role list
# /root/.ansible/roles
- geerlingguy.docker, 7.2.0
```

To install a role from a local repository directory ...

```
root@a47952ea7696:/go# lax collection install --server=/tmp/foo geerlingguy.mac
...
root@a47952ea7696:/go# ansible-galaxy collection list
# /root/.ansible/collections/ansible_collections
Collection      Version
--------------- -------
geerlingguy.mac 4.0.1
```

If you decided to create an http repository server or found one on the internet, swap the /tmp/foo from the previous examples with the url to the server's location with the repository path. ie `--server=https://tannerjc.net/galaxy`
