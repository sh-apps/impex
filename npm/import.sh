set -ex

for x in *.tgz
do
				npm publish $x --registry=http://localhost:8081/repository/npm/ --access=public
done
