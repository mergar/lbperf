#!/bin/sh
SOURCE_DIR="$1"
MAX_RECORDS="100"
DT=$( date )
echo "${DT}: wakeup: ${SOURCE_DIR}" >> /tmp/log.txt

DATA_ROOT_DST="/usr/local/www/nginx/data"

TR_CMD="tr"

# generic mandatory tools/script
MAIN_CMD="
tr
cat
chmod
cut
sed
sort
find
grep
rm
jq
ln
rmdir
basename
readlink
realpath
date
uname
mv
mkdir
"

for i in ${MAIN_CMD}; do
        mycmd=
        mycmd=$( which ${i} )
        if [ ! -x "${mycmd}" ]; then
                echo "${pgm} error: no such executable dependency/requirement: ${i}" | tee -a /tmp/wakeup.log
                exit 1
        fi
        MY_CMD=$( echo ${i} | ${TR_CMD} '\-[:lower:]' '_[:upper:]' )
        MY_CMD="${MY_CMD}_CMD"
        eval "${MY_CMD}=\"${mycmd}\""
done

[ ! -d "${DATA_ROOT_DST}" ] && ${MKDIR_CMD} -p ${DATA_ROOT_DST}

## обрабатываем SOURCE_DIR
if [ 3 -gt 2 ]; then

if [ ! -d "${SOURCE_DIR}" ]; then
	echo "no such source dir: ${SOURCE_DIR}"
	exit 0
fi

SQLITE3_CMD=$( command -v sqlite3 )
if [ -z "${SQLITE3_CMD}" ]; then
    echo "no such sqlite3" >&2
    exit 1
fi
TAR_CMD=$( command -v tar )
if [ -z "${TAR_CMD}" ]; then
    echo "no such tar" >&2
    exit 1
fi

last_identify_path=$( find ${SOURCE_DIR}/ -type f -name identify.json.* | sort -V | tail -n1 )

if [ -z "${last_identify_path}" ]; then
	echo "no such identify file"
	exit 0
fi
if [ ! -r "${SOURCE_DIR}/data.tgz" ]; then
	echo "no such data.tgz"
	exit 0
fi

# basename
TOKEN="${SOURCE_DIR##*/}"

# basename
last_identify="${last_identify_path##*/}"

#echo $last_identify
tid="${last_identify##*.}"

if [ -z "${tid}" ]; then
	echo "no tid from identify file"
	exit 0
fi

OSNAME=$( uname -s )

case "${OSNAME}" in
	FreeBSD)
		formatted_date=$( date -r "$tid" +%Y-%m )
		;;
	*)
		formatted_date=$( date -d @"$tid" +%Y-%m )
		;;
esac


DB_FILE="/tmp/data.${formatted_date}.sqlite"
OPWD=$( pwd )
cd ${SOURCE_DIR}
${TAR_CMD} xfz data.tgz

${RM_CMD} ${SOURCE_DIR}/data.tgz ${SOURCE_DIR}/${last_identify}

CONFIG_FILE="${SOURCE_DIR}/tests/config"

# 2. Проверка наличия файла конфигурации
if [ ! -f "${CONFIG_FILE}" ]; then
	echo "Ошибка: Файл конфигурации ${CONFIG_FILE} не найден." >&2
	exit 1
fi

# 3. Импорт переменных в текущую оболочку POSIX shell
. "${CONFIG_FILE}"

${MV_CMD} ${SOURCE_DIR}/tests ${SOURCE_DIR}/tests.${tid}

if [ ! -d ${DATA_ROOT_DST}/${TOKEN} ]; then
	${CHMOD_CMD} 0755 ${SOURCE_DIR}
	${MV_CMD} ${SOURCE_DIR} ${DATA_ROOT_DST}/
else
	${MV_CMD} ${SOURCE_DIR}/* ${DATA_ROOT_DST}/${TOKEN}/
fi

cd ${DATA_ROOT_DST}
echo "${LN_CMD} -sf ${TOKEN}/tests.${tid} ${tid}"
${LN_CMD} -sf ${TOKEN}/tests.${tid} ${tid}
ret=$?
if [ ${ret} -eq 0 ]; then
	${CAT_CMD} >${DATA_ROOT_DST}/${tid}/index.html <<________EOF
<html>
<body>
<img src="graph.svg"></img>

<pre>
________EOF

${GREP_CMD} -v ^progress ${DATA_ROOT_DST}/${tid}/work/*.log >> ${DATA_ROOT_DST}/${tid}/index.html

	${CAT_CMD} >>${DATA_ROOT_DST}/${tid}/index.html <<________EOF
</pre>
</body>
</html>
________EOF
fi

# 5. Проверка существования таблицы и её создание (CREATE TABLE IF NOT EXISTS)
# Все текстовые поля создаются как TEXT, числовые — как NUMERIC
${SQLITE3_CMD} "$DB_FILE" <<EOF
CREATE TABLE IF NOT EXISTS h2load (
    tid TEXT PRIMARY KEY,
    CPU TEXT,
    BACKEND TEXT,
    BODY_PAYLOAD_SIZE NUMERIC,
    BACKEND_SOCKET TEXT,
    FRONTEND_PROTO TEXT,
    BACKEND_PROTO TEXT,
    CONCURRENT TEXT,
    THREAD NUMERIC,
    REQUESTS_NUM NUMERIC,
    FULL_TEST_TIME NUMERIC,
    RPS_MAX NUMERIC,
    RPS_AVG NUMERIC,
    CONCURRENT_CLIENTS NUMERIC,
    time NUMERIC,
    rps_float NUMERIC,
    bw NUMERIC,
    requests_total NUMERIC,
    requests_started NUMERIC,
    requests_done NUMERIC,
    requests_succeeded NUMERIC,
    codes_2xx NUMERIC,
    codes_3xx NUMERIC,
    codes_4xx NUMERIC,
    codes_5xx NUMERIC,
    codes_failed NUMERIC,
    codes_errored NUMERIC,
    codes_timeout NUMERIC,
    traffic_total_bytes NUMERIC,
    traffic_headers_bytes NUMERIC,
    traffic_data_bytes NUMERIC,
    request_min NUMERIC,
    request_max NUMERIC,
    request_median NUMERIC,
    request_p95 NUMERIC,
    request_p99 NUMERIC,
    request_mean NUMERIC,
    request_sd NUMERIC,
    request_plus_minus_sd NUMERIC,
    connect_min NUMERIC,
    connect_max NUMERIC,
    connect_median NUMERIC,
    connect_p95 NUMERIC,
    connect_p99 NUMERIC,
    connect_mean NUMERIC,
    connect_sd NUMERIC,
    connect_plus_minus_sd NUMERIC,
    ttfb_min NUMERIC,
    ttfb_max NUMERIC,
    ttfb_median NUMERIC,
    ttfb_p95 NUMERIC,
    ttfb_p99 NUMERIC,
    ttfb_mean NUMERIC,
    ttfb_sd NUMERIC,
    ttfb_plus_minus_sd NUMERIC,
    req_per_s_min NUMERIC,
    req_per_s_max NUMERIC,
    req_per_s_median NUMERIC,
    req_per_s_p95 NUMERIC,
    req_per_s_p99 NUMERIC,
    req_per_s_mean NUMERIC,
    req_per_s_sd NUMERIC,
    req_per_s_plus_minus_sd NUMERIC
);
EOF

# 6. Вставка или обновление данных (INSERT OR REPLACE)
# Используются двойные кавычки во внешнем heredoc для раскрытия переменных shell
${SQLITE3_CMD} "$DB_FILE" <<EOF
INSERT OR REPLACE INTO h2load (
    tid, CPU, BACKEND, BODY_PAYLOAD_SIZE, BACKEND_SOCKET, FRONTEND_PROTO, BACKEND_PROTO,
    CONCURRENT, THREAD, REQUESTS_NUM, FULL_TEST_TIME, RPS_MAX, RPS_AVG, CONCURRENT_CLIENTS, time,
    rps_float, bw, requests_total, requests_started, requests_done, requests_succeeded,
    codes_2xx, codes_3xx, codes_4xx, codes_5xx, codes_failed, codes_errored, codes_timeout,
    traffic_total_bytes, traffic_headers_bytes, traffic_data_bytes, request_min, request_max,
    request_median, request_p95, request_p99, request_mean, request_sd, request_plus_minus_sd,
    connect_min, connect_max, connect_median, connect_p95, connect_p99, connect_mean,
    connect_sd, connect_plus_minus_sd, ttfb_min, ttfb_max, ttfb_median, ttfb_p95,
    ttfb_p99, ttfb_mean, ttfb_sd, ttfb_plus_minus_sd, req_per_s_min, req_per_s_max,
    req_per_s_median, req_per_s_p95, req_per_s_p99, req_per_s_mean, req_per_s_sd,
    req_per_s_plus_minus_sd
) VALUES (
    '$tid', '$CPU', '$BACKEND', '$BODY_PAYLOAD_SIZE', '$BACKEND_SOCKET', '$FRONTEND_PROTO', '$BACKEND_PROTO',
    '$CONCURRENT', '$THREAD', '$REQUESTS_NUM', '$FULL_TEST_TIME', '$RPS_MAX', '$RPS_AVG', '$CONCURRENT_CLIENTS', '$time',
    '$rps', '$bw', '$requests_total', '$requests_started', '$requests_done', '$requests_succeeded',
    '$codes_2xx', '$codes_3xx', '$codes_4xx', '$codes_5xx', '$codes_failed', '$codes_errored', '$codes_timeout',
    '$traffic_total_bytes', '$traffic_headers_bytes', '$traffic_data_bytes', '$request_min', '$request_max',
    '$request_median', '$request_p95', '$request_p99', '$request_mean', '$request_sd', '$request_plus_minus_sd',
    '$connect_min', '$connect_max', '$connect_median', '$connect_p95', '$connect_p99', '$connect_mean',
    '$connect_sd', '$connect_plus_minus_sd', '$ttfb_min', '$ttfb_max', '$ttfb_median', '$ttfb_p95',
    '$ttfb_p99', '$ttfb_mean', '$ttfb_sd', '$ttfb_plus_minus_sd', '$req_per_s_min', '$req_per_s_max',
    '$req_per_s_median', '$req_per_s_p95', '$req_per_s_p99', '$req_per_s_mean', '$req_per_s_sd',
    '$req_per_s_plus_minus_sd'
);
EOF

ret=$?

if [ ${ret} -eq 0 ]; then
	echo "inserted/updated: tid=$tid"
else
	echo "failed"
	exit 1
fi

# END OF обрабатываем SOURCE_DIR
fi


## Начало генерации HTML
echo "HTML gen"
cat /root/send-data/modules/lb-perf/html-header.html > /usr/local/www/nginx/index.html

${FIND_CMD} ${DATA_ROOT_DST}/ -type d -name tests.\* -exec ${REALPATH_CMD} {} \; | while read _line; do
        basename=$( ${BASENAME_CMD} ${_line} )
        #echo $basename
        ts=$( echo ${basename} | ${SED_CMD} -e "s:tests.::g")
        echo "${ts} ${_line}"
done | ${SORT_CMD} -n -k1 -r > /tmp/sort

cur_item_num=0

${CAT_CMD} /tmp/sort | while read _ts _data; do
        echo "rec: ${cur_item_num}/${MAX_RECORDS}"
        if [ ${cur_item_num} -gt ${MAX_RECORDS} ]; then
                echo "prune oldest recods: ${_data}"
#                ${RM_CMD} -rf ${_data}
                continue
        fi

        if [ ! -r ${_data}/config ]; then
                echo "No such ${_data}/config" >> /tmp/log.log
                continue
        fi

#        ${LN_CMD} -sf ${_data}/${bs} ${progdir}/../index/${_ts}

        . ${_data}/config

        _human_date=$( ${DATE_CMD} -r "${_ts}" "+%Y-%m-%d %H:%M:%S" )

        echo "create svg:" >> /tmp/log.log

        echo "/root/newbench/scripts/create-gnuplot.sh ${_data} graph.svg"
        /root/newbench/scripts/create-gnuplot.sh ${_data} graph.svg >> /tmp/log.log 2>&1
        #svg_link=$( echo ${_data}/FreeBSD-tests.${_ts}-chart.svg | sed 's:/usr/local/www/nginx-dist::g' )
	svg_link="/data/${_ts}/graph.svg"
	#svg_link="/data/${_ts}"

cat >>/usr/local/www/nginx/index.html<<EOF
                                    <tr class="table-area-object" data-id="${_ts}">
                                        <td data-id="${_ts}" class="table-area-property check-multy">
                                            <input type="checkbox" name="compare" value="${_ts}" onchange="setSelections()">
                                        </td>
                                        <td class="table-area-property">#${_ts}<br>(${_human_date}):<br> ${LAB_DESC}</td>
                                        <td class="table-area-property">виртуальная среда?: ${is_virtual}, ${CPU}</td>
                                        <td class="table-area-property">backend: ${BACKEND}, FRONTEND PROTO: ${FRONTEND_PROTO}, BACKEND_PROTO:${BACKEND_PROTO}, 
lab_nginx_tls:${lab_nginx_tls}, 
lab_haproxy_tls:${lab_haproxy_tls}, 
lab_nginx_listen:${lab_nginx_listen}, 
lab_haproxy_ports:${lab_haproxy_ports}, 
</td>


<!-- профиль нагрузки -->

                                        <td class="table-area-property">runtime=${FULL_TEST_TIME}, THREAD: ${THREAD}, REQUESTS_NUM: ${REQUESTS_NUM}, CONCURRENT: ${CONCURRENT}, 
BODY_PAYLOAD_SIZE:${BODY_PAYLOAD_SIZE} 
                                        </td>
<!-- результат -->
                                        <td class="table-area-property">

<table style="border: none;">
  <tr>
    <td>RPS_AVG ${RPS_AVG}</td>
    <td>RPS_MAX ${RPS_MAX}</td>
    <td>BW: ${bw}</td>
    <td>&nbsp;</td>
    <td>&nbsp;</td>
  </tr>
  <tr>
    <td>&nbsp;</td>
    <td>&nbsp;</td>
    <td>&nbsp;</td>
    <td>&nbsp;</td>
    <td>&nbsp;</td>
  </tr>
  <tr style="background-color: #f2f2f2; font-weight: bold;">
    <td style="font-weight: bold;">&nbsp;</td>
    <td>max</td>
    <td>median</td>
    <td>p95</td>
    <td>p99</td>
  </tr>
  <tr>
    <td style="background-color: #f2f2f2; font-weight: bold;">request (us)</td>
    <td style="text-decoration: underline;">$request_max</td>
    <td style="text-decoration: underline;">$request_median</td>
    <td style="text-decoration: underline;">$request_p95</td>
    <td style="text-decoration: underline;">$request_p99</td>
  </tr>
  <tr>
    <td style="background-color: #f2f2f2; font-weight: bold;">connect (us)</td>
    <td style="text-decoration: underline;">$connect_max</td>
    <td style="text-decoration: underline;">$connect_median</td>
    <td style="text-decoration: underline;">$connect_p95</td>
    <td style="text-decoration: underline;">$connect_p99</td>
  </tr>
  <tr>
    <td style="background-color: #f2f2f2; font-weight: bold;">TTFB (us)</td>
    <td style="text-decoration: underline;">$ttfb_max</td>
    <td style="text-decoration: underline;">$ttfb_median</td>
    <td style="text-decoration: underline;">$ttfb_p95</td>
    <td style="text-decoration: underline;">$ttfb_p99</td>
  </tr>
  <tr>
    <td style="background-color: #f2f2f2; font-weight: bold;">req/s</td>
    <td style="text-decoration: underline;">$req_per_s_max</td>
    <td style="text-decoration: underline;">$req_per_s_median</td>
    <td style="text-decoration: underline;">$req_per_s_p95</td>
    <td style="text-decoration: underline;">$req_per_s_p99</td>
  </tr>
</table>


</td>

<!-- SVG -->
                                        <td class="table-area-property">
<a href="/data/${_ts}" target="_blank" title="svg chart">
    <img src="${svg_link}" width="320" height="240" alt="svg chart" style="object-fit: contain;">
</a>
                                    </tr>
EOF

done

cat /root/send-data/modules/lb-perf/html-footer.html >> /usr/local/www/nginx/index.html

exit 0
