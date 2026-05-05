ALTER TABLE "option"
  ADD UNIQUE ("key");

---- create above / drop below ----
ALTER TABLE "option"
  DROP CONSTRAINT "option_key_key";
