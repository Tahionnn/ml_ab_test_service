#!/bin/sh
set -e

BOOTSTRAP="kafka:29092"

create_topic () {
  kafka-topics --bootstrap-server $BOOTSTRAP \
    --create --if-not-exists \
    --topic $1 \
    --partitions 1 \
    --replication-factor 1
}

create_topic prediction.events
create_topic deployment.requested
create_topic deployment.artifact.download.requested
create_topic deployment.artifact.downloaded
create_topic deployment.started
create_topic deployment.ready
create_topic deployment.failed
create_topic deployment.undeploy.requested
create_topic deployment.undeployed