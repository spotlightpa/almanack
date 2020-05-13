CREATE TABLE site_data (
  id serial PRIMARY KEY,
  key text NOT NULL UNIQUE,
  data jsonb NOT NULL DEFAULT '{}' ::jsonb,
  created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER row_updated_at_on_site_data_trigger_
  BEFORE UPDATE ON site_data
  FOR EACH ROW
  EXECUTE PROCEDURE update_row_updated_at_function_ ();

INSERT INTO site_data (KEY, data)
  VALUES ('data/editorsPicks.json', '
{
  "featuredStories": [],
  "limitSubfeatures": true,
  "subfeatures": [],
  "subfeaturesLimit": 2,
  "topSlots": []
}');

---- create above / drop below ----
DROP TABLE site_data;
