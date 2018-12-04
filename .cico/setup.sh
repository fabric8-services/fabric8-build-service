#!/bin/bash
#
# Build script for CI builds on CentOS CI
set -ex

export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

REPO_PATH=${GOPATH}/src/github.com/fabric8-services/fabric8-build
REGISTRY="quay.io"

function setup() {
    if [ -f jenkins-env.json ]; then
        eval "$(./env-toolkit load -f jenkins-env.json \
                ghprbActualCommit \
                ghprbPullAuthorLogin \
                ghprbGhRepository \
                ghprbPullId \
                GIT_COMMIT \
                FABRIC8_DEVCLUSTER_TOKEN \
                DEVSHIFT_TAG_LEN \
                QUAY_USERNAME \
                QUAY_PASSWORD \
                BUILD_URL \
                BUILD_ID)"
    fi

    # We need to disable selinux for now, XXX
    /usr/sbin/setenforce 0 || :

    yum install epel-release --enablerepo=extras -y
    yum -y install --enablerepo=epel-testing buildah podman make golang git

    mkdir -p $(dirname ${REPO_PATH})
    cp -a ${HOME}/payload ${REPO_PATH}
    cd ${REPO_PATH}

    echo 'CICO: Build environment created.'
}

function tag_push() {
    local image="$1"
    local tag="$2"

    buildah tag ${image}:latest ${image}:${tag}
    buildah push ${image}:${tag} ${image}:${tag}
}

function deploy() {
  # Login first
  cd ${REPO_PATH}

  if [ -n "${QUAY_USERNAME}" -a -n "${QUAY_PASSWORD}" ]; then
      podman login -u ${QUAY_USERNAME} -p ${QUAY_PASSWORD} ${REGISTRY}
  else
      echo "Could not login, missing credentials for the registry"
      exit 1
  fi

  # Build fabric8-build
  make image

  TAG=$(echo $GIT_COMMIT | cut -c1-${DEVSHIFT_TAG_LEN})
  if [ "$TARGET" = "rhel" ]; then
    tag_push ${REGISTRY}/openshiftio/rhel-fabric8-services-fabric8-build $TAG
    tag_push ${REGISTRY}/openshiftio/rhel-fabric8-services-fabric8-build latest
  else
    tag_push ${REGISTRY}/openshiftio/fabric8-services-fabric8-build $TAG
    tag_push ${REGISTRY}/openshiftio/fabric8-services-fabric8-build latest
  fi

  echo 'CICO: Image pushed, ready to update deployed app'
}

function check_up() {
    service=$1
    host=$2
    port=$3
    max=30 # 1 minute

    counter=1
    while true;do
        python -c "import socket;s = socket.socket(socket.AF_INET, socket.SOCK_STREAM);s.connect(('$host', $port))" \
        >/dev/null 2>/dev/null && break || \
        echo "CICO: Waiting that $service on ${host}:${port} is started (sleeping for 2)"

        if [[ ${counter} == ${max} ]];then
            echo "CICO: Could not connect to ${service} after some time"
            echo "CICO: Investigate locally the logs with fig logs"
            exit 1
        fi

        sleep 2

        (( counter++ ))
    done
}

function run_postgres() {
    make container-run

    eval $(make print-env|grep '^DB_CONTAINER_PORT')
    export F8_POSTGRES_PORT="${DB_CONTAINER_PORT}"

    check_up postgres-build 127.0.0.1 ${DB_CONTAINER_PORT}

	./bin/fabric8-build -migrateDatabase
}

function compile() {
    make build
}

function do_coverage() {
    make coverage

    # Upload to codecov
    bash <(curl -s https://codecov.io/bash) -K -X search -f tmp/coverage.out -t 533b56c6-9fec-4ff2-9756-6aea46d46f2b
}

function download_latest_oc() {
    pushd /bin >/dev/null && \
curl -s -L $(curl -L -s "https://api.github.com/repos/openshift/origin/releases/latest"|python -c "import sys, json;x=json.load(sys.stdin);print([ r['browser_download_url'] for r in x['assets'] if 'openshift-origin-client-tools' in r['name'] and 'linux-64bit' in r['name']][0])") -o /tmp/oc.tgz && \
    tar xz -f/tmp/oc.tgz --wildcards "*/oc" --strip-components=1 && \
popd >/dev/null
}

function deploy_devcluster() {
    project="${1}"

    download_latest_oc || { yum install centos-release-openshift-origin && yum install origin-clients ;}

    oc login --insecure-skip-tls-verify=true https://devtools-dev.ext.devshift.net:8443 --token=${FABRIC8_DEVCLUSTER_TOKEN}

    exist=$(oc get project -o go-template='{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'|grep "^${project}$" || true)
    if [[ -z ${exist} ]];then
        oc new-project ${project}
    fi
    oc project ${project}
    bash ./openshift/deploy-openshift-dev.sh
}

function do_test() {
    make test-unit
}
