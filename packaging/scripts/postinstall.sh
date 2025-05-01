#!/bin/sh

set -e

CONFIG_FILE=/etc/yggops/config.yaml

mkdir -p /var/lib/yggops /etc/yggops
chown -R yggops:yggops /var/lib/yggops /etc/yggops /usr/share/yggops

if [ ! -f "$CONFIG_FILE" ]; then
  cp /usr/share/yggops/default_config.yaml $CONFIG_FILE
  chown yggops:yggops "$CONFIG_FILE"
fi

systemctl daemon-reload
systemctl unmask yggops.service
systemctl enable yggops.service
systemctl start yggops.service
