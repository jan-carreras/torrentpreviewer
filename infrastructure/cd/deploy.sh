#!/usr/bin/env bash

runuser -l torrent -c "cd /opt/torrentpreview/src && make build"

systemctl restart tphttp && sleep 1 && systemctl status --no-pager -l tphttp
echo
systemctl restart tpevents && sleep 1 && systemctl status --no-pager -l tpevents
