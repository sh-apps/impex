#!/bin/bash
for x in `cat dependencies.txt`; do apt-cache depends -i $x | awk '/Depends:/ {print $2}' | xargs  apt-get download; apt-get download $x; done

#for x in `ls *.deb`; do curl -u "$USERNAME:$PASSWORD" -H "Content-Type: multipart/form-data" --data-binary "@./$x" "https://${NEXUS_HOST}:${NEXUS_PORT}/repository/hosted-apt/"; done
