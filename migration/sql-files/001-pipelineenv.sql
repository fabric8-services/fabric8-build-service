CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE pipelines (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text UNIQUE NOT NULL,
    space_id uuid,
    PRIMARY KEY(id)
);


CREATE TABLE environments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    environment_id uuid NOT NULL,
    pipeline_id uuid,
    FOREIGN KEY (pipeline_id) REFERENCES pipelines(id),
    PRIMARY KEY(id)
);

CREATE INDEX environments_environment_id_idx ON environments USING BTREE (environment_id);
CREATE INDEX pipelines_space_id_idx ON pipelines USING BTREE (space_id);
CREATE INDEX environments_pipeline_id_idx ON environments USING BTREE (pipeline_id);
