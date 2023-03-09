#!/bin/bash
#Takes all config and append to binary => a way to have configs+binary versioned together and always available...
path=$1
binary=$2
echo "Ready to patch $2 with files in  directory $1"
#find $1 -type f \( -name "*.json" -o -name "*.toml" \)
printf "\n#NOMAD#\n" >> $binary

#Useless... (without it, can be read easily with 2 split)
#printf "#NOMAD_JSON#\n" >> $binary

json=""
while read -r appConfig; do
	echo "Adding $appConfig"
	app=$(tr -d "[:blank:]\t\r\n" < "$appConfig")
	if [ "${json}" = "" ];then
	  json="${app}"
	else
	  json="${json},${app}"
	fi
done <<<$(find $path -maxdepth 1 -name "*.json")

echo "{\"apps\":[${json}]}" >> $binary

printf "#NOMAD_TOML#\ntitle=\"nomad\"\n[apps]\n" >> $binary
# shellcheck disable=SC2086
find $path -maxdepth 1 -type f -name "*.toml" | xargs cat | tr -d "[:blank:]\t" >> "$binary"
#Show added apps
grep "\[apps\..*\]" "$binary" | while read -r appConfig; do
	echo "Added $appConfig"
done
printf "\n" >> "$binary"