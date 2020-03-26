CREATE TABLE domain_roles (
    id serial PRIMARY KEY,
    domain text NOT NULL,
    roles text[],
    created_at timestamp WITH time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp WITH time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX unique_domain_on_domain_roles ON domain_roles ((lower("domain")) text_ops);

CREATE OR REPLACE FUNCTION update_row_updated_at_function_ ()
    RETURNS TRIGGER
    AS $$
BEGIN
    NEW.updated_at = clock_timestamp();
    RETURN NEW;
END;
$$
LANGUAGE 'plpgsql';

CREATE TRIGGER row_updated_at_on_domain_roles_trigger_
    BEFORE UPDATE ON domain_roles
    FOR EACH ROW
    EXECUTE PROCEDURE update_row_updated_at_function_ ();

INSERT INTO "domain_roles" ("domain", "roles")
    VALUES
        -- Fixtures
        ('spotlightpa.org', '{"Spotlight PA","arc user",editor}'),
        --
        ('inquirer.com', '{"arc user",editor}');

---- create above / drop below ----
DROP TABLE domain_roles;

