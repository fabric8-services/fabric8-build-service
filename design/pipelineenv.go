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

var pipelineEnvMap = a.Type("PipelineEnvironmentMaps", func() {
	a.Description(`JSONAPI store for data of pipeline environments.`)
	a.Attribute("id", d.UUID, "ID of the pipeline environment map", func() {
		a.Example("40bbdd3d-8b5d-4fd6-ac90-7236b669af04")
	})
	a.Attribute("spaceID", d.UUID, "ID of the pipeline environment map", func() {
		a.Example("40bbdd3d-8b5d-4fd6-ac90-7236b669af04")
	})
	a.Attribute("name", d.String, "The environment name", func() {
		a.Example("myapp-stage")
	})
	a.Attribute("environments", a.ArrayOf(envAttrs), "An array of environments")
	a.Attribute("links", genericLinks)
	a.Required("name", "environments")
})

var pipelineEnvMapListMeta = a.Type("PipelineEnvironmentListMeta", func() {
	a.Attribute("totalCount", d.Integer)
	a.Required("totalCount")
})

var pipelineEnvMapSingle = JSONSingle(
	"PipelineEnvironmentMap", "Holds a single pipeline environment map",
	pipelineEnvMap,
	nil)

var pipelineEnvMapList = JSONList(
	"PipelineEnvironmentMaps", "Holds the list of pipeline environment map",
	pipelineEnvMap,
	pagingLinks,
	pipelineEnvMapListMeta)

var _ = a.Resource("PipelineEnvironmentMaps", func() {
	a.Action("create", func() {
		a.Description("Create pipeline environment map")
		a.Params(func() {
			a.Param("spaceID", d.UUID, "Space ID for the pipeline environment map")
		})
		a.Routing(
			a.POST("/spaces/:spaceID/pipeline-environment-maps"),
		)
		a.Payload(pipelineEnvMapSingle)
		a.Response(d.Created, pipelineEnvMapSingle)
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
			a.GET("/spaces/:spaceID/pipeline-environment-maps"),
		)
		a.Response(d.OK, pipelineEnvMapList)
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
			a.GET("/pipeline-environment-maps/:ID"),
		)
		a.Response(d.OK, pipelineEnvMapSingle)
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
			a.PATCH("/pipeline-environment-maps/:ID"),
		)
		a.Payload(pipelineEnvMapSingle)
		a.Response(d.OK, pipelineEnvMapSingle)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.NotFound, JSONAPIErrors)
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
		a.Response(d.MethodNotAllowed, JSONAPIErrors)
		a.Response(d.Conflict, JSONAPIErrors)
		a.Response(d.Forbidden, JSONAPIErrors)
	})

})
