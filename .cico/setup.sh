#!/bin/bash
#
# Build script for CI builds on CentOS CI
set -ex

export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

REPO_PATH=${GOPATH}/src/github.com/fabric8-services/fabric8-build-service
REGISTRY="quay.io"

function addCommentToPullRequest() {
    message="$1"
    pr="$2"
    project="$3"
    url="https://api.github.com/repos/${project}/issues/${pr}/comments"

    set +x
    echo curl -X POST -s -L -H "Authorization: XXXX|base64 --decode)" ${url} -d "{\"body\": \"${message}\"}"
    curl -X POST -s -L -H "Authorization: token $(echo ${FABRIC8_HUB_TOKEN}|base64 --decode)" ${url} -d "{\"body\": \"${message}\"}"
    set -x
}

function setup() {
    if [ -f jenkins-env.json ]; then
        eval "$(./env-toolkit load -f jenkins-env.json \
                FABRIC8_HUB_TOKEN \
                ghprbActualCommit \
                ghprbPullAuthorLogin \
                ghprbGhRepository \
                ghprbPullId \
                GIT_COMMIT \
                QUAY_USERNAME \
                QUAY_PASSWORD \
                BUILD_URL \
                BUILD_ID)"
    fi

    # We need to disable selinux for now, XXX
    /usr/sbin/setenforce 0 || :

    yum -y install docker make golang git
    service docker start

    mkdir -p $(dirname ${REPO_PATH})
    cp -a ${HOME}/payload ${REPO_PATH}
    cd ${REPO}

    echo 'CICO: Build environment created.'
}

function tag_push() {
    local image="$1"
    local tag="$2"

    docker tag ${image}:latest ${image}:${tag}
    docker push ${image}:${tag}
}


function _deploy() {
  # Login first
  cd ${REPO_PATH}


  if [ -n "${QUAY_USERNAME}" -a -n "${QUAY_PASSWORD}" ]; then
    docker login -u ${QUAY_USERNAME} -p ${QUAY_PASSWORD} ${REGISTRY}
  else
    echo "Could not login, missing credentials for the registry"
  fi

  # Build fabric8-build-service
  make image

  TAG=$(echo $GIT_COMMIT | cut -c1-${DEVSHIFT_TAG_LEN})
  if [ "$TARGET" = "rhel" ]; then
    tag_push ${REGISTRY}/openshiftio/rhel-fabric8-services-fabric8-build-service $TAG
    tag_push ${REGISTRY}/openshiftio/rhel-fabric8-services-fabric8-build-service latest
  else
    tag_push ${REGISTRY}/openshiftio/fabric8-services-fabric8-build-service $TAG
    tag_push ${REGISTRY}/openshiftio/fabric8-services-fabric8-build-service latest
  fi

  echo 'CICO: Image pushed, ready to update deployed app'
}

function deploy() {
    set +e
    _deploy || fail=true
    set -e

    if [[ -n ${fail} ]];then
        addCommentToPullRequest "Merge job has failed see: ${BUILD_URL}/consoleText" "${ghprbPullId}" "${ghprbGhRepository}"
        exit 1
    fi
}

function dotest() {
    cd ${REPO_PATH}

    make build
    make test-unit

    make analyze-go-code
    make coverage

    # Upload to codecov
    bash <(curl -s https://codecov.io/bash) -K -X search -f tmp/coverage.out -t 533b56c6-9fec-4ff2-9756-6aea46d46f2b
}
