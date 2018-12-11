#!/bin/bash
# Deploy a fully working fabric8-build on an openshift cluster
#
# You can do this directly on minishift or on a remote cluster
#
# TODO(chmouel): add an option to not clean up DB datas
# This is for dev, everyting get cleaned up again and again
#
# Configuration is inside makefile

set -ex

function readlinkf() { python -c 'import os,sys;print(os.path.realpath(sys.argv[1]))' $1 ;}
cd $(dirname $(readlinkf $0))/../

eval $(make print-env|egrep '^(REGISTRY_URL|WIT|AUTH|DB|ENV).*(IMAGE|PORT|NAME)')

oc whoami 2>/dev/null >/dev/null || { echo "oc does not seem to be configured properly"; exit 1 ;}

SERVER=$(oc project|sed 's/.*on server.*https:\/\///;s/:.*//')
FORCE_DELETE_VARS="--force --wait=true --grace-period=0"
DC_DB=db

IMAGE_PULL_POLICY="IfNotPresent"
[[ ${SERVER} =~ .*(devshift.net|openshift.com) ]] && {
    IMAGE_PULL_POLICY="Always"
}

function waitForDC() {
    DC=$1
    local max=60 # 2mn
    local cnt=1

    while [[ $(oc get dc/${DC} -o json|python -c "import sys, json;x=json.load(sys.stdin);print x['status']['availableReplicas']") < 1 ]];do
        [[ ${cnt} > ${max} ]] && {
            echo "Timing out while waiting for DC/${DC}";
            exit 1
        }

        sleep 2
        (( cnt++ ))
    done
}

function deploy_db() {
    DBS=${1}
    POSTGRESQL_ADMIN_PASSWORD=`sed -n '/postgres.password/ { s/.*: //;p ;}' config.yaml`
    # Let's make sure we delete everything
    oc delete is -l app=db
    oc delete dc -l app=db --cascade=true 2>/dev/null || true
    oc new-app --name=${DC_DB} -e \
       POSTGRESQL_ADMIN_PASSWORD=${POSTGRESQL_ADMIN_PASSWORD} \
       ${DB_CONTAINER_IMAGE} -o yaml | oc delete ${FORCE_DELETE_VARS} -f- 2>/dev/null || true
    sleep 2
    oc new-app --name=${DC_DB} -e \
       POSTGRESQL_ADMIN_PASSWORD=${POSTGRESQL_ADMIN_PASSWORD} \
       ${DB_CONTAINER_IMAGE}

    waitForDC ${DC_DB}

    sleep 2

    for x in ${DBS};do
        cnt=1
        while true;do
            [[ ${cnt} -ge 50 ]] && { echo "Cannot connect to database"; exit 1 ; }
            oc rsh dc/${DC_DB} psql -c  "create database ${x};" && break || {
                (( cnt++ ))
                sleep 5
           }
        done
    done

}

function deploy_sideservice() {
    local name=${1}
    local image=${2}
    local env_dbs="${3}"

    oc delete is -l app=${name} 2>/dev/null || true
    oc new-app --name="${name}" ${image} -o yaml | \
        oc delete ${FORCE_DELETE_VARS} -f- 2>/dev/null || true
    sleep 2
    oc new-app --name="${name}" ${env_dbs} ${image}
    sleep 2
    oc delete route/${name} || true
    oc expose service/${name}
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
       F8_POSTGRES_HOST=${DC_DB} \
       F8_POSTGRES_DATABASE=build \
       F8_DEVELOPER_MODE_ENABLED=1 \
       F8_ENVIRONMENT=local F8_POSTGRES_SSLMODE=disable \
       F8_AUTH_URL="http://${AUTH_CONTAINER_NAME}:${AUTH_CONTAINER_PORT}"

    oc delete ${FORCE_DELETE_VARS} route/f8build 2>/dev/null || true
    oc expose service/f8build
}

deploy_db "build auth env"

# AUTH
AUTH_SERVICE_VARIABLES=$(cat <<EOF
-e AUTH_LOG_LEVEL=debug
-e AUTH_POSTGRES_HOST=${DC_DB}
-e AUTH_POSTGRES_DATABASE=auth
-e AUTH_POSTGRES_PORT=5432
-e AUTH_DEVELOPER_MODE_ENABLED=true
EOF
)
deploy_sideservice ${AUTH_CONTAINER_NAME} ${AUTH_CONTAINER_IMAGE} "${AUTH_SERVICE_VARIABLES}"

# ENV
ENV_SERVICE_VARIABLES=$(cat <<EOF
-e F8_LOG_LEVEL=debug
-e F8_POSTGRES_HOST=${DC_DB}
-e F8_POSTGRES_DATABASE=env
-e F8_POSTGRES_PORT=5432
-e F8_DEVELOPER_MODE_ENABLED=true
-e F8_AUTH_URL=http://${AUTH_CONTAINER_NAME}:${AUTH_CONTAINER_PORT}
EOF
)
deploy_sideservice ${ENV_CONTAINER_NAME} ${ENV_CONTAINER_IMAGE} "${ENV_SERVICE_VARIABLES}"

# WIT
WIT_SERVICE_VARIABLES=$(cat <<EOF
-e F8_LOG_LEVEL=debug
-e F8_POSTGRES_HOST=${DC_DB}
-e F8_POSTGRES_DATABASE=wit
-e F8_POSTGRES_PORT=5432
-e F8_DEVELOPER_MODE_ENABLED=true
-e F8_AUTH_URL=http://${AUTH_CONTAINER_NAME}:${AUTH_CONTAINER_PORT}
EOF
)
deploy_sideservice ${WIT_CONTAINER_NAME} ${WIT_CONTAINER_IMAGE} "${WIT_SERVICE_VARIABLES}"


# Build
deploy_app
