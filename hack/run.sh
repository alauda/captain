#!/usr/bin/env sh

cd /captain

file="/tmp/repositories.yaml"
dir="/captain/.helm/repository/"
if [[ -f "$file" ]]
then
	echo "$file found."
	mkdir -p ${dir}
	cp ${file} ${dir}
else
	echo "$file not found, waiting for captain to create"
fi

exec "$@"