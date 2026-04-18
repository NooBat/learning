-- flux_dev schema for L01
-- One table: workflows. Execution, tenancy, and more complex constraints
-- arrive in later levels (L02 auth-tenancy, L05 async-work). Keep this
-- file small and hand-written at L01; a migration tool lands at L03.
--
-- Design note on `trigger_type`: enforced at the DB layer via CHECK,
-- not via ENUM. Posture — trigger_type is an evolving business concept,
-- not a platonic type. We expect to add values (cron, event, file-drop,
-- upstream-complete) and possibly retire some of the initial set as the
-- workflow model matures. CHECK makes value add/rename/remove a routine
-- DROP/ADD constraint migration; ENUM would punish value removal
-- (rename-type + migrate + drop is a full rebuild). Revisit at L03 when
-- proper migration tooling arrives, or sooner if the taxonomy stabilizes
-- — promotion path to ENUM is a single ALTER COLUMN TYPE.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";  -- provides gen_random_uuid()

CREATE TABLE IF NOT EXISTS workflows (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    name          text        NOT NULL,
    trigger_type  text        NOT NULL CHECK (trigger_type IN ('schedule', 'webhook', 'manual')),
    steps         jsonb       NOT NULL DEFAULT '[]'::jsonb,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);
