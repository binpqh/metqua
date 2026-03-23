using SimpleRag.Core;

namespace SimpleRag.Storage;

/// <summary>
/// In-memory vector store using cosine similarity
/// NOTE: This is a simple implementation. For production, consider FAISS or Qdrant.
/// </summary>
public class VectorStore
{
    private readonly List<VectorEntry> _vectors;
    private readonly Dictionary<string, Chunk> _chunks;
    private readonly int _dimension;
    private int _nextIndex = 0;

    public VectorStore(int dimension = 768)
    {
        _dimension = dimension;
        _vectors = new List<VectorEntry>();
        _chunks = new Dictionary<string, Chunk>();
    }

    /// <summary>
    /// Add a chunk with its embedding to the store
    /// </summary>
    public void AddChunk(Chunk chunk)
    {
        if (chunk.Embedding == null)
        {
            throw new ArgumentException("Chunk must have an embedding", nameof(chunk));
        }

        if (chunk.Embedding.Length != _dimension)
        {
            throw new ArgumentException(
                $"Embedding dimension mismatch. Expected {_dimension}, got {chunk.Embedding.Length}",
                nameof(chunk));
        }

        var entry = new VectorEntry
        {
            Index = _nextIndex,
            Vector = chunk.Embedding,
            ChunkId = chunk.Id
        };

        chunk.VectorIndex = _nextIndex;
        _vectors.Add(entry);
        _chunks[chunk.Id] = chunk;
        _nextIndex++;
    }

    /// <summary>
    /// Add multiple chunks at once
    /// </summary>
    public void AddChunks(IEnumerable<Chunk> chunks)
    {
        foreach (var chunk in chunks)
        {
            AddChunk(chunk);
        }
    }

    /// <summary>
    /// Search for top K most similar chunks to a query vector
    /// </summary>
    public List<SearchResult> SearchTopK(float[] queryVector, int k = 5)
    {
        if (queryVector.Length != _dimension)
        {
            throw new ArgumentException(
                $"Query vector dimension mismatch. Expected {_dimension}, got {queryVector.Length}");
        }

        if (_vectors.Count == 0)
        {
            return new List<SearchResult>();
        }

        // Calculate cosine similarity for all vectors
        var results = _vectors
            .Select(entry => new
            {
                Entry = entry,
                Similarity = CosineSimilarity(queryVector, entry.Vector)
            })
            .OrderByDescending(x => x.Similarity)
            .Take(k)
            .Select(x => new SearchResult
            {
                Chunk = _chunks[x.Entry.ChunkId],
                Score = x.Similarity,
                Index = x.Entry.Index
            })
            .ToList();

        return results;
    }

    /// <summary>
    /// Get a chunk by its ID
    /// </summary>
    public Chunk? GetChunk(string chunkId)
    {
        return _chunks.TryGetValue(chunkId, out var chunk) ? chunk : null;
    }

    /// <summary>
    /// Get all chunks
    /// </summary>
    public List<Chunk> GetAllChunks()
    {
        return _chunks.Values.ToList();
    }

    /// <summary>
    /// Clear all data from the store
    /// </summary>
    public void Clear()
    {
        _vectors.Clear();
        _chunks.Clear();
        _nextIndex = 0;
    }

    /// <summary>
    /// Get statistics about the vector store
    /// </summary>
    public VectorStoreStats GetStats()
    {
        return new VectorStoreStats
        {
            TotalVectors = _vectors.Count,
            TotalChunks = _chunks.Count,
            Dimension = _dimension,
            UniqueDocuments = _chunks.Values.Select(c => c.SourceId).Distinct().Count()
        };
    }

    /// <summary>
    /// Calculate cosine similarity between two vectors
    /// </summary>
    private static float CosineSimilarity(float[] a, float[] b)
    {
        if (a.Length != b.Length)
        {
            throw new ArgumentException("Vectors must have the same dimension");
        }

        float dotProduct = 0f;
        float magnitudeA = 0f;
        float magnitudeB = 0f;

        for (int i = 0; i < a.Length; i++)
        {
            dotProduct += a[i] * b[i];
            magnitudeA += a[i] * a[i];
            magnitudeB += b[i] * b[i];
        }

        magnitudeA = MathF.Sqrt(magnitudeA);
        magnitudeB = MathF.Sqrt(magnitudeB);

        if (magnitudeA == 0 || magnitudeB == 0)
        {
            return 0f;
        }

        return dotProduct / (magnitudeA * magnitudeB);
    }

    #region Nested Types

    private class VectorEntry
    {
        public int Index { get; set; }
        public float[] Vector { get; set; } = Array.Empty<float>();
        public string ChunkId { get; set; } = string.Empty;
    }

    #endregion
}

/// <summary>
/// Result from a vector search
/// </summary>
public class SearchResult
{
    public Chunk Chunk { get; set; } = null!;
    public float Score { get; set; }
    public int Index { get; set; }
}

/// <summary>
/// Statistics about the vector store
/// </summary>
public class VectorStoreStats
{
    public int TotalVectors { get; set; }
    public int TotalChunks { get; set; }
    public int Dimension { get; set; }
    public int UniqueDocuments { get; set; }
}
