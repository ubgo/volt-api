## Vector partial index issue

- https://github.com/pgvector/pgvector-go/issues/27
  https://github.com/pgvector/pgvector?tab=readme-ov-file#can-i-store-vectors-with-different-dimensions-in-the-same-column

```sql
CREATE INDEX "item_embedding_1536" ON "items" USING hnsw ((embedding::vector(1536))  vector_l2_ops) WHERE vec_dims=1536;
CREATE INDEX "item_embedding_768" ON "items" USING hnsw ((embedding::vector(768))  vector_l2_ops) WHERE vec_dims=768;
```
