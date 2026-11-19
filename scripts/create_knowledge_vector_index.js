// Creates the Atlas Vector Search index used by POST /movie (vector search over knowledge_chunks).
//
// Run once against the Atlas cluster, after embeddings have been generated
// (AI layer: PATCH /ai/embeddings):
//   mongosh "<MONGO_URI>" scripts/create_knowledge_vector_index.js
//
// The index name must match VECTOR_SEARCH_INDEX (default: knowledge_chunks_vector_index).
// numDimensions is read from an existing embedding, so it always matches the embedding model.

const INDEX_NAME = "knowledge_chunks_vector_index";

const kb = db.getSiblingDB("ticket-booking-system");
const sample = kb.knowledge_chunks.findOne({ "embedding.0": { $exists: true } });
if (!sample) {
  throw new Error(
    "No knowledge_chunks document has a non-empty embedding. Generate embeddings first (PATCH /ai/embeddings)."
  );
}
const numDimensions = sample.embedding.length;
print(`Creating vector index '${INDEX_NAME}' with ${numDimensions} dimensions`);

kb.knowledge_chunks.createSearchIndex({
  name: INDEX_NAME,
  type: "vectorSearch",
  definition: {
    fields: [
      { type: "vector", path: "embedding", numDimensions: numDimensions, similarity: "cosine" },
      { type: "filter", path: "source_id" },
    ],
  },
});

print("Index requested. Atlas builds it in the background; check its status in the Atlas UI or with:");
print(`  db.getSiblingDB("ticket-booking-system").knowledge_chunks.getSearchIndexes("${INDEX_NAME}")`);
