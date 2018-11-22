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

var pipelineEnvSingle = JSONSingle(
	"PipelineEnvironment", "Holds a single pipeline environment map",
	pipelineEnv,
	nil)

var _ = a.Resource("PipelineEnvironments", func() {
	a.Action("create", func() {
		a.Routing(
			a.POST("/pipelines/environments/:spaceID"),
		)
		a.Description("Create environment")
		a.Params(func() {
			a.Param("spaceID", d.UUID, "UUID of the space")
		})
		a.Payload(pipelineEnvSingle)
		a.Response(d.Created, pipelineEnvSingle)
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
		a.Response(d.MethodNotAllowed, JSONAPIErrors)
		a.Response(d.Conflict, JSONAPIErrors)
	})

	a.Action("show", func() {
		a.Description("Retrieve pipeline environment map (as JSONAPI) for the given space ID.")
		a.Params(func() {
			a.Param("spaceID", d.UUID, "Space ID for the pipeline environment map")
		})

		a.Routing(
			a.GET("/pipeline/environments/:spaceID"),
		)

		a.Params(func() {
			a.Param("spaceID", d.UUID, "UUID of the space")
		})

		a.Response(d.OK, pipelineEnvSingle)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.NotFound, JSONAPIErrors)
	})

})
