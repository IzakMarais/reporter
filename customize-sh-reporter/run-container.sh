#!/usr/bin/env bash

# Copyright © 2021.  SmartHub Inc. All rights reserved. This product is protected by copyright
# and intellectual property laws in the United States and other countries as well as by international treaties.

SERVICE_NAME=grafana-pdf-reporter
IOTC_SERVICE_NAME=insights-$SERVICE_NAME
IOTC_SERVICE_ARTIFACTS=/opt/smarthub/iotc/$IOTC_SERVICE_NAME
IOTC_SERVICE_LOGS_FOLDER=/var/log/smarthub/iotc
IOTC_CONFIG_FOLDER=/opt/smarthub/iotc/conf


# Starting the service as container

# --env-file $IOTC_CONFIG_FOLDER/grafana-pdf-reporter.env \
echo "Starting the $SERVICE_NAME service container..."
docker run -d \
    --expose 8686 \
    --name=$IOTC_SERVICE_NAME \
    --hostname=$IOTC_SERVICE_NAME \
    --network "host" \
    --workdir $IOTC_SERVICE_ARTIFACTS \
      ghcr.io/smarthub-ai/shinternal/grafana-pdf-reporter:latest \
    -grid-layout=1 \
    -templates=$IOTC_SERVICE_ARTIFACTS
docker cp logo.png $IOTC_SERVICE_NAME:/$IOTC_SERVICE_ARTIFACTS
docker cp customTemplate.tex $IOTC_SERVICE_NAME:/$IOTC_SERVICE_ARTIFACTS