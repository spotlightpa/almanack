CREATE TABLE image (
    id serial PRIMARY KEY,
    path text NOT NULL UNIQUE,
    description text NOT NULL DEFAULT '',
    credit text NOT NULL DEFAULT '',
    src_url text NOT NULL DEFAULT '',
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER row_updated_at_on_image_trigger_
    BEFORE UPDATE ON image
    FOR EACH ROW
    EXECUTE PROCEDURE update_row_updated_at_function_ ();

---- create above / drop below ----
DROP TABLE image;

