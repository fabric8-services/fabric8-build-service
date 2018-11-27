Fabric8-build
=============

Fabric8 build service is the build service providing a REST API for the build operations on [OpenShift.IO](https://openshift.io)

API
---

Swagger documentation available [here](http://swagger.goa.design/?url=github.com%2Ffabric8-services%2Ffabric8-build%2Fdesign#/).

Development
-----------

### Deployment

There is a easy way to deploy the whole service and all its dependencies on a OpenShift environment, as long you have enough resources and a oc cli cluster access, you can run the script :

[openshift/deploy-openshift-dev.sh](openshift/deploy-openshift-dev.sh)

and this will deploy the services :

* auth
* env
* current build

with all the configuration needed,
The script is not idempotent it will just clean itself on multiple runs,


### Building

#### Prerequisites

go (>=1.8), git, make, dep

#### Make commands

* If you run `make` without argument it would print all the useful target to run with a help.
* `make build` will build fabric8-build
* `make regenerate` will regenerate the goa generated field
* `make container-run` will start a DB container into docker
* `make coverage` will run the coverage (need DB)
* `make test-unit` will run all the test units  (need DB)
* `make analyze-go-code` will run the [golangci](https://github.com/golangci/golangci-lint) static analysis tool, run it before sending your PR to github. The [golangci](https://golangci.com/r/github.com/fabric8-services/fabric8-build) check service will be the one taking care of checking your PR, so it can report nicely on your code if there is any issues.

### RUNNING

* You need to make sure your database is populated, you can do it like so :

```bash
make migrate-database
```

This will run all the sql files from [migration/sql-files/*](migration/sql-files/)

* You need to have a auth server in the variable F8_AUTH_URL, if you were using the [openshift/deploy-openshift-dev.sh](openshift/deploy-openshift-dev.sh) script you can get the public route and expose it directly like this :

```bash
`export F8_AUTH_URL="http://$(oc get route auth -o json|jq -r .spec.host)"`
```

* You may want to add those debug variables :

```shell
export F8_LOG_LEVEL=debug;
export F8_DEVELOPER_MODE_ENABLED=1;
export F8_ENABLE_DB_LOGS=1
```

* Just start with `./bin/fabric8-build`

* You can use `make dev` to automate this and have it restarted and recompiled when there is a refresh of the code, this is using a [forked](https://github.com/chmouel/fresh/) version of [fresh](https://github.com/pilu/fresh/).

### Testing functional

* You would need a token you can generate a token from a dev fabric8auth service like this :

```shell
export TOKEN=$(curl -s -L ${F8_AUTH_URL}/api/token/generate -H 'content-type: application/json' | jq -r '.[0].token.access_token')
```

and then use it in your curl command, for example to create a pipeline environment :

```shell
curl -v -H "Authorization: Bearer ${TOKEN}" -H 'Content-Type: application/vndpipelineenvironment' -d@/tmp/example.json http://localhost:8080/api/pipelines/environments/$(uuid)
```

### Unit tests

* You can run all unit tests with `make test unit`, by default it will run quietly if you need to increase verbosity for debugging you can do like this

```bash
GO_TEST_VERBOSITY_FLAG="-v" F8_LOG_LEVEL="debug" make test-unit
```

* If you want to run individual unit test you can do like this :

```bash
fabric8-build/build $ F8_ENABLE_DB_LOGS=1 F8_POSTGRES_PORT=5433 F8_LOG_LEVEL=1 F8_DEVELOPER_MODE_ENABLED=1 F8_RESOURCE_DATABASE=1 F8_RESOURCE_UNIT_TEST=1 go test -v -run TestEnvironmentController
```

In this case `TestEnvironmentController` is the TestSuite which would run the whole things, make sure you have the other environment variable and adjust if needed (i.e: the `POSTGRES PORT`)

AUTHORS
-------

Fabric8-build-team <devtools-build@redhat.com>
