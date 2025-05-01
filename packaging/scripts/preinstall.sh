#!/bin/sh

set -e

if ! getent group yggops >/dev/null; then
  groupadd --system yggops
fi

if ! id -u yggops >/dev/null 2>&1; then
  useradd --system --gid yggops --home /var/lib/yggops --shell /sbin/nologin yggops
fi
