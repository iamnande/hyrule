CREATE ROLE hyrule_app_ro LOGIN PASSWORD 'hyrule_app_ro';

GRANT CONNECT ON DATABASE hyrule TO hyrule_app_ro;
GRANT USAGE ON SCHEMA public TO hyrule_app_ro;

ALTER DEFAULT PRIVILEGES FOR ROLE hyrule_owner IN SCHEMA public
    GRANT SELECT ON TABLES TO hyrule_app_ro;
