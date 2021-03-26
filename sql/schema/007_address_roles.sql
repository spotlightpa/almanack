CREATE TABLE address_roles (
  id serial PRIMARY KEY,
  email_address text NOT NULL,
  roles text[],
  created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX unique_email_address_on_address_roles ON address_roles
  ((lower("email_address")) text_ops);

CREATE TRIGGER row_updated_at_on_address_roles_trigger_
  BEFORE UPDATE ON address_roles
  FOR EACH ROW
  EXECUTE PROCEDURE update_row_updated_at_function_ ();

---- create above / drop below ----
DROP TABLE address_roles;
