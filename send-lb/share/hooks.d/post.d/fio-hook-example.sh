#!/bin/sh

# save result into custom file in /tmp/fio directory:
[ ! -d /tmp/fio ] && mkdir /tmp/fio
cat > /tmp/fio/${PROFILE}-${DEVICE}.txt <<EOF
DEVICE: ${DEVICE}
PROFILE: ${PROFILE}
BS: ${BS}
DEPTH: ${DEPTH}
JOBS: ${JOBS}
RWMIXREAD: ${RWMIXREAD}
FS: ${FS}
R_MB: ${R_MB}
R_KB: ${R_KB}
W_MB: ${W_MB}
W_KB: ${W_KB}
R_IOPS: ${R_IOPS}
R_IOPS_MAX: ${R_IOPS_MAX}
R_IOPS_MIN: ${R_IOPS_MIN}
W_IOPS: ${W_IOPS}
W_IOPS_MAX: ${W_IOPS_MAX}
W_IOPS_MIN: ${W_IOPS_MIN}
EOF

# show file content
cat /tmp/fio/${PROFILE}-${DEVICE}.txt
