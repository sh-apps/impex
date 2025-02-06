# Impex

Download all the stuff we need for airgapped environments from an input file.

## Tasks

### download-npm

```
go run *.go npm export -lock-file=../app-nodejs/package-lock.json
```

### download-vsix

```
go run *.go vsix -file=./vsix.txt
```

### download-containers

```
go run *.go container -file=./containers.txt
```

### download-actions

This requires the https://github.com/actions/actions-sync tool.

```
mkdir -p package/actions
actions-sync pull --repo-name-list-file=actions.txt --cache-dir=package/actions
```

### download-deb-packages

Populate deb/dependencies.txt with a list of top-level deb packages you want.

```
mkdir -p package/deb
cd deb
sh download_debs.sh
mv *.deb ../package/deb/
