-- +goose Up
-- Where an item is in its lifecycle (see spec/domain-model.md). Milestone 02
-- implements 'planned' <-> 'packed'; 'prepared', 'needs_buying' and 'bought'
-- arrive with the rest of the chain.
--
-- Deliberately no CHECK on the allowed values: the set grows in a later
-- milestone, and changing a CHECK in SQLite means rebuilding the table. The
-- handlers decide which transitions are allowed; at two users a stray value
-- is not the risk worth paying a table rebuild to prevent.
ALTER TABLE item ADD COLUMN status TEXT NOT NULL DEFAULT 'planned';

-- +goose Down
ALTER TABLE item DROP COLUMN status;
