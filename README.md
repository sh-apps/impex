# Impex

Download all the stuff we need for airgapped environments from an input file.

## Tasks

### download-npm

```
go run *.go npm -lock-file=../app-nodejs/package-lock.json
```

### download-vsix

```
go run *.go vsix -file=./vsix.txt
```

### download-containers

```
go run *.go container -file=./containers.txt
```

