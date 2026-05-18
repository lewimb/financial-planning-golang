-- Fix amount column type: DECIMAL(10,2) is inconsistent with Go domain's `int` type
-- and IDR currency (no decimal subunits). BIGINT aligns with TransactionRequest.Amount int.
ALTER TABLE transactions ALTER COLUMN amount TYPE BIGINT USING amount::BIGINT;
