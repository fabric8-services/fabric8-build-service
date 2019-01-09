CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE pipeline_env_maps (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    space_id uuid,
    PRIMARY KEY(id),
    UNIQUE(name, space_id)
);


CREATE TABLE pipeline_environments (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    environment_id uuid NOT NULL,
    pipelineenvmap_id uuid,
    FOREIGN KEY (pipelineenvmap_id) REFERENCES pipeline_env_maps(id)
);

CREATE INDEX pipeline_environments_environment_id_idx ON pipeline_environments USING BTREE (environment_id);
CREATE INDEX pipeline_env_maps_space_id_idx ON pipeline_env_maps USING BTREE (space_id);
CREATE INDEX pipeline_environments_pipelineenvmap_id_idx ON pipeline_environments USING BTREE (pipelineenvmap_id);