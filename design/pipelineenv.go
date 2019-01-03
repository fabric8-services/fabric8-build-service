package design

import (
	d "github.com/goadesign/goa/design"
	a "github.com/goadesign/goa/design/apidsl"
)

var envAttrs = a.Type("EnvironmentAttributes", func() {
	a.Description(`JSONAPI store for the environment UUID.`)
	a.Attribute("envUUID", d.UUID, "UUID of the environment", func() {
		a.Example("40bbdd3d-8b5d-4fd6-ac90-7236b669af04")
	})
})

var pipelineEnv = a.Type("PipelineEnvironments", func() {
	a.Description(`JSONAPI store for data of pipeline environments.`)
	a.Attribute("id", d.UUID, "ID of the pipeline environment", func() {
		a.Example("40bbdd3d-8b5d-4fd6-ac90-7236b669af04")
	})
	a.Attribute("spaceID", d.UUID, "ID of the pipeline environment", func() {
		a.Example("40bbdd3d-8b5d-4fd6-ac90-7236b669af04")
	})
	a.Attribute("name", d.String, "The environment name", func() {
		a.Example("myapp-stage")
	})
	a.Attribute("environments", a.ArrayOf(envAttrs), "An array of environMents")
	a.Attribute("links", genericLinks)
	a.Required("name", "environments")
})

var pipelineEnvListMeta = a.Type("PipelineEnvironmentListMeta", func() {
	a.Attribute("totalCount", d.Integer)
	a.Required("totalCount")
})

var pipelineEnvSingle = JSONSingle(
	"PipelineEnvironment", "Holds a single pipeline environment map",
	pipelineEnv,
	nil)

var pipelineEnvList = JSONList(
	"PipelineEnvironments", "Holds the list of pipeline environment map",
	pipelineEnv,
	pagingLinks,
	pipelineEnvListMeta)

var _ = a.Resource("PipelineEnvironments", func() {
	a.Action("create", func() {
		a.Description("Create pipeline environment map")
		a.Params(func() {
			a.Param("spaceID", d.UUID, "Space ID for the pipeline environment map")
		})
		a.Routing(
			a.POST("/spaces/:spaceID/pipeline-environments"),
		)
		a.Payload(pipelineEnvSingle)
		a.Response(d.Created, pipelineEnvSingle)
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
		a.Response(d.MethodNotAllowed, JSONAPIErrors)
		a.Response(d.Forbidden, JSONAPIErrors)
		a.Response(d.Conflict, JSONAPIErrors)
		a.Response(d.NotFound, JSONAPIErrors)
	})

	a.Action("list", func() {
		a.Description("Retrieve list of pipeline environment maps (as JSONAPI) for the given space ID.")
		a.Params(func() {
			a.Param("spaceID", d.UUID, "Space ID for the pipeline environment map")
		})
		a.Routing(
			a.GET("/spaces/:spaceID/pipeline-environments"),
		)
		a.Response(d.OK, pipelineEnvList)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
		a.Response(d.Forbidden, JSONAPIErrors)
	})

	a.Action("show", func() {
		a.Description("Retrieve pipeline environment map (as JSONAPI) for the given ID.")
		a.Params(func() {
			a.Param("ID", d.UUID, "ID of the pipeline environment map")
		})
		a.Routing(
			a.GET("/pipeline-environments/:ID"),
		)
		a.Response(d.OK, pipelineEnvSingle)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.NotFound, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
		a.Response(d.Forbidden, JSONAPIErrors)
	})

	a.Action("update", func() {
		a.Description("Update the pipeline environment map for the given ID.")
		a.Params(func() {
			a.Param("ID", d.UUID, "ID of the pipeline environment map to update")
		})
		a.Routing(
			a.PATCH("/pipeline-environments/:ID"),
		)
		a.Payload(pipelineEnvSingle)
		a.Response(d.OK, pipelineEnvSingle)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.NotFound, JSONAPIErrors)
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
		a.Response(d.MethodNotAllowed, JSONAPIErrors)
		a.Response(d.Conflict, JSONAPIErrors)
		a.Response(d.Forbidden, JSONAPIErrors)
	})

})
