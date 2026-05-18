-- goals.deleted_at is dead schema: goalRepository.Delete() executes
-- `DELETE FROM goals WHERE id=$1 AND user_id=$2` (hard delete).
-- No goal query filters on deleted_at. Drop the column to align schema with behavior.
ALTER TABLE goals DROP COLUMN IF EXISTS deleted_at;
