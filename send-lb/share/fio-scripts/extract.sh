#!/bin/sh
pgm="${0##*/}"          # Program basename
progdir="${0%/*}"       # Program directory

export PATH="/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin"

# generic mandatory tools/script
TR_CMD=$( which tr )
if [ -z "${TR_CMD}" ]; then
	echo "no such command: tr"
	exit 1
fi

MAIN_CMD="
cat
tr
jq
"
for i in ${MAIN_CMD}; do
	mycmd=
	mycmd=$( which ${i} )
	if [ ! -x "${mycmd}" ]; then
		echo "${pgm} error: no such executable dependency/requirement: ${i}"
		exit 1
	fi
	MY_CMD=$( echo ${i} | ${TR_CMD} '\-[:lower:]' '_[:upper:]' )
	MY_CMD="${MY_CMD}_CMD"
	eval "${MY_CMD}=\"${mycmd}\""
done

if [ ! -r ${progdir}/../fio/tests/4k/randread-1-2.json ]; then
	echo "No such result: ${progdir}/../fio/tests/4k/randread-1-2.json"
	exit 1
else
	echo "Using: ${progdir}/../fio/tests/4k/randread-1-2.json"
fi
READ_BW_BYTES=$( ${CAT_CMD} ${progdir}/../fio/tests/4k/randread-1-2.json | ${JQ_CMD} '.jobs[0].read.bw_bytes' | ${TR_CMD} -d '"' )
WRITE_BW_BYTES=$( ${CAT_CMD} ${progdir}/../fio/tests/4k/randread-1-2.json | ${JQ_CMD} '.jobs[0].write.bw_bytes' | ${TR_CMD} -d '"' )
RUNTIME=$( ${CAT_CMD} ${progdir}/../fio/tests/4k/randread-1-2.json | ${JQ_CMD} '.jobs[0]."job options".runtime' | ${TR_CMD} -d '"' )

echo $READ_BW_BYTES
echo $WRITE_BW_BYTES
echo $RUNTIME

MB=$(( READ_BW_BYTES / 1024 / 1024 ))
KB=$(( READ_BW_BYTES / 1024 ))

echo "MB per sec: $MB"
echo "KB per sec: $KB"

#MB_PER_SEC=$(( RUN_TIME / MB ))
#echo $MB_PER_SEC


#Run status group 0 (all jobs):
#   READ: bw=45.9MiB/s (48.1MB/s), 22.6MiB/s-23.2MiB/s (23.7MB/s-24.3MB/s), io=2752MiB (2885MB), run=60001-60001msec
