-- Reports table had only type+generated_at; insufficient for any real report feature.
-- Add month/year for period scoping, content for report body, status for generation state.
ALTER TABLE reports
    ADD COLUMN month   INTEGER,
    ADD COLUMN year    INTEGER,
    ADD COLUMN content TEXT,
    ADD COLUMN status  VARCHAR(20) DEFAULT 'GENERATED';
