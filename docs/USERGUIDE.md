# User guide

## Preface

LAX's primary workflow is to install .tar.gz files from a http or file location. There is not yet a central host for lax, so it is the user's responsibility currently to make their own repo to host the content. However, making a repo is far easier than it sounds.

A LAX repo, aka "server", is a directory with two subdirectories (roles & collections). In the root directory there is a json file named `repometa.json`. LAX inspects the repometa.json to figure out what other index files exist for the repo (usually role_manfiests.tar.gz & collection_manifests.tar.gz). The index files provide LAX with all the information it needs about the available content and their dependencies.

When the user installs a package from the LAX repo, it finds the necessary package and dependencies from the index files then downloads the relevant tar.gz files and extracts them to the appropriate location for `ansible` or `ansible-playbook` to use.

## Installing LAX

## Syncing Content

## Creating a Repo

## Hosting a Repo

## Installing Content From a Repo
