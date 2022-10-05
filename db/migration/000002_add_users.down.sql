-- Drop owner_currency_key constrain
ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTS "owner_currency_key";
-- drop the foreign key constraint for the owner field of account table
ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTS "accounts_owner_fkey";
-- drop user table
DROP TABLE IF EXISTS "users";