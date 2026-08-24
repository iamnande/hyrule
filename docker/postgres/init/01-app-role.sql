-- the service connects as this role, never as the POSTGRES_USER above.
-- postgres silently bypasses RLS for the table owner and superusers unless
-- FORCE ROW LEVEL SECURITY is set - connecting as a non-owner role from day
-- one means that trap never gets a chance to bite. see
-- docs/conventions.md#data-layer.
CREATE ROLE hyrule_app LOGIN PASSWORD 'hyrule_app';

GRANT CONNECT ON DATABASE hyrule TO hyrule_app;
GRANT USAGE ON SCHEMA public TO hyrule_app;

-- covers tables/sequences created by migrations (which run as the owner
-- role, POSTGRES_USER) after this script has already run once.
ALTER DEFAULT PRIVILEGES FOR ROLE hyrule_owner IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO hyrule_app;
ALTER DEFAULT PRIVILEGES FOR ROLE hyrule_owner IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO hyrule_app;
