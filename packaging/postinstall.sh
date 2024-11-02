#!/bin/sh

systemctl daemon-reload
systemctl enable gombus2mqtt.service
systemctl restart gombus2mqtt.service
