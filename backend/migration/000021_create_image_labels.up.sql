CREATE TABLE image_labels (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    image_id   UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    label      TEXT NOT NULL,
    score      FLOAT4 NOT NULL
);

CREATE INDEX ON image_labels (image_id);

INSERT INTO image_labels (image_id, label, score)
SELECT images.id,
       elem->>'Description',
       (elem->>'Score')::float4
FROM images,
     LATERAL jsonb_array_elements(ai_labels) AS elem
WHERE ai_labels IS NOT NULL
  AND jsonb_typeof(ai_labels) = 'array';
