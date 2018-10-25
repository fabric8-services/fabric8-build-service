#!/bin/bash
set -ex

function readlinkf() { python -c 'import os,sys;print(os.path.realpath(sys.argv[1]))' $1 ;}
cd $(dirname $(readlinkf $0))/../

eval $(make print-env|egrep '^(REGISTRY_URL|AUTH|DB).*(IMAGE|PORT|NAME)')

oc whoami 2>/dev/null >/dev/null || { echo "oc does not seem to be configured properly"; exit 1 ;}

SERVER=$(oc project|sed 's/.*on server.*https:\/\///;s/:.*//')

IMAGE_PULL_POLICY="IfNotPresent"
[[ ${SERVER} =~ .*(devshift.net|openshift.com) ]] && {
    IMAGE_PULL_POLICY="Always"
}

function deploy_db() {
    POSTGRESQL_ADMIN_PASSWORD=`sed -n '/postgres.password/ { s/.*: //;p ;}' config.yaml`
    oc new-app --name=${1} -e \
       POSTGRESQL_ADMIN_PASSWORD=${POSTGRESQL_ADMIN_PASSWORD} \
       ${DB_CONTAINER_IMAGE} -o yaml | oc delete -f- 2>/dev/null || true
    sleep 2
    oc new-app --name=${1} -e \
       POSTGRESQL_ADMIN_PASSWORD=${POSTGRESQL_ADMIN_PASSWORD} \
       ${DB_CONTAINER_IMAGE}
}

function deploy_auth() {
    oc new-app --name="${AUTH_CONTAINER_NAME}" ${AUTH_CONTAINER_IMAGE} -o yaml | \
        oc delete -f- 2>/dev/null || true
    sleep 2
    oc new-app --name="${AUTH_CONTAINER_NAME}" -e AUTH_POSTGRES_HOST="${AUTH_DB_CONTAINER_NAME}" -e AUTH_POSTGRES_PORT=5432 \
       -e AUTH_DEVELOPER_MODE_ENABLED=true ${AUTH_CONTAINER_IMAGE}
}

function deploy_app() {
    for i in config app;do
        oc process -f openshift/f8build.${i}.yaml|oc delete -f- 2>/dev/null || true
    done
    sleep 2

    # HACK TO REMOVE WHEN WE HAVE https://gitlab.cee.redhat.com/dtsd/housekeeping/issues/2406 merged
    oc process -f openshift/f8build.app.yaml IMAGE=${REGISTRY_URL_IMAGE} -o yaml | \
        sed -e "s/imagePullPolicy: Always/imagePullPolicy: ${IMAGE_PULL_POLICY}/" | \
        oc create -f-

    # Make it after so configChange kicks in
    oc process -f openshift/f8build.config.yaml |
        oc create -f-

    # make it after applying the configs
    oc set env dc/f8build \
       F8_LOG_LEVEL=debug \
       F8_DEVELOPER_MODE_ENABLED=1 \
       F8_ENVIRONMENT=local F8_POSTGRES_SSLMODE=disable \
       F8_AUTH_URL="http://${AUTH_CONTAINER_NAME}:${AUTH_CONTAINER_PORT}"

    oc delete route/f8build 2>/dev/null || true
    oc expose service/f8build
}

deploy_db db

deploy_db ${AUTH_DB_CONTAINER_NAME}

deploy_auth

deploy_app
