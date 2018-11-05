#!/bin/bash
# Deploy a fully working fabric8-build on an openshift cluster
#
# Configuration is inside makefile

set -ex

function readlinkf() { python -c 'import os,sys;print(os.path.realpath(sys.argv[1]))' $1 ;}
cd $(dirname $(readlinkf $0))/../

eval $(make print-env|egrep '^(REGISTRY_URL|AUTH|DB|ENV).*(IMAGE|PORT|NAME)')

oc whoami 2>/dev/null >/dev/null || { echo "oc does not seem to be configured properly"; exit 1 ;}

SERVER=$(oc project|sed 's/.*on server.*https:\/\///;s/:.*//')

FORCE_DELETE_VARS="--force --wait=true --grace-period=0"

IMAGE_PULL_POLICY="IfNotPresent"
[[ ${SERVER} =~ .*(devshift.net|openshift.com) ]] && {
    IMAGE_PULL_POLICY="Always"
}

function deploy_db() {
    POSTGRESQL_ADMIN_PASSWORD=`sed -n '/postgres.password/ { s/.*: //;p ;}' config.yaml`
    oc new-app --name=${1} -e \
       POSTGRESQL_ADMIN_PASSWORD=${POSTGRESQL_ADMIN_PASSWORD} \
       ${DB_CONTAINER_IMAGE} -o yaml | oc delete ${FORCE_DELETE_VARS} -f- 2>/dev/null || true
    sleep 2
    oc new-app --name=${1} -e \
       POSTGRESQL_ADMIN_PASSWORD=${POSTGRESQL_ADMIN_PASSWORD} \
       ${DB_CONTAINER_IMAGE}
}

function deploy_sideservice() {
    local name=${1}
    local image=${2}
    local env_dbs="${3}"

    oc new-app --name="${name}" ${image} -o yaml | \
        oc delete ${FORCE_DELETE_VARS} -f- 2>/dev/null || true
    sleep 2
    oc new-app --name="${name}" ${env_dbs} -e AUTH_DEVELOPER_MODE_ENABLED=true \
       ${image}
}

function deploy_app() {
    for i in config app;do
        oc process -f openshift/f8build.${i}.yaml|oc delete ${FORCE_DELETE_VARS} -f- 2>/dev/null || true
    done

    sleep 2

    # Apply the openshift templates
    oc process -f openshift/f8build.app.yaml IMAGE=${REGISTRY_URL_IMAGE} -o yaml | \
        sed -e "s/imagePullPolicy: Always/imagePullPolicy: ${IMAGE_PULL_POLICY}/" | \
        oc create -f-

    # Apply configs and secrets
    oc process -f openshift/f8build.config.yaml | oc create -f-

    # do this after applying the configs
    oc set env dc/f8build \
       F8_LOG_LEVEL=debug \
       F8_DEVELOPER_MODE_ENABLED=1 \
       F8_ENVIRONMENT=local F8_POSTGRES_SSLMODE=disable \
       F8_AUTH_URL="http://${AUTH_CONTAINER_NAME}:${AUTH_CONTAINER_PORT}"

    oc delete ${FORCE_DELETE_VARS} route/f8build 2>/dev/null || true
    oc expose service/f8build
}

deploy_db db
deploy_db ${AUTH_DB_CONTAINER_NAME}
deploy_db ${ENV_DB_CONTAINER_NAME}

AUTH_SERVICE_VARIABLES=$(cat <<EOF
-e AUTH_LOG_LEVEL=debug
-e AUTH_POSTGRES_HOST=${AUTH_DB_CONTAINER_NAME}
-e AUTH_POSTGRES_PORT=5432
EOF
)
deploy_sideservice ${AUTH_CONTAINER_NAME} ${AUTH_CONTAINER_IMAGE} "${AUTH_SERVICE_VARIABLES}"

ENV_SERVICE_VARIABLES=$(cat <<EOF
-e F8_LOG_LEVEL=debug
-e F8_POSTGRES_HOST=${ENV_DB_CONTAINER_NAME}
-e F8_POSTGRES_PORT=5432
-e F8_AUTH_URL=http://${AUTH_CONTAINER_NAME}:${AUTH_CONTAINER_PORT}
EOF
)
deploy_sideservice ${ENV_CONTAINER_NAME} ${ENV_CONTAINER_IMAGE} "${ENV_SERVICE_VARIABLES}"
oc expose service/${ENV_CONTAINER_NAME}

deploy_app
