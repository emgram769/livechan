#!/usr/bin/env bash
#
# add channels manually
#
#cd $(dirname $0)
vals=""
for arg in $@ ; do
    vals="(\"$arg\"),$vals"
done
sqlite3 livechan.db <<EOF
INSERT INTO Channels(name) VALUES${vals:0:-1};
EOF
