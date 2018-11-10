CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE pipelines (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text,
    space_id uuid
);


CREATE TABLE environments (
    id uuid primary key DEFAULT uuid_generate_v4() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    environment_id uuid NOT NULL,
    pipeline_id uuid NOT NULL
);

CREATE INDEX environments_environment_id_idx ON environments USING BTREE (environment_id);
CREATE INDEX environments_pipeline_id_idx ON environments USING BTREE (pipeline_id);
CREATE INDEX pipelines_space_id_idx ON pipelines USING BTREE (space_id);
