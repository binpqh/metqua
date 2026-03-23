namespace SimpleRag.Core;

/// <summary>
/// Represents a chunk of text from a document with its embedding
/// </summary>
public class Chunk
{
    /// <summary>
    /// Unique identifier for the chunk (format: documentId_chunkIndex)
    /// </summary>
    public string Id { get; set; } = string.Empty;

    /// <summary>
    /// The text content of this chunk
    /// </summary>
    public string Text { get; set; } = string.Empty;

    /// <summary>
    /// Approximate token count (word-based estimation)
    /// </summary>
    public int Tokens { get; set; }

    /// <summary>
    /// ID of the source document
    /// </summary>
    public string SourceId { get; set; } = string.Empty;

    /// <summary>
    /// Source file path for reference
    /// </summary>
    public string SourcePath { get; set; } = string.Empty;

    /// <summary>
    /// Position (word offset) in the original document
    /// </summary>
    public int Position { get; set; }

    /// <summary>
    /// Vector embedding for this chunk (null until embedded)
    /// </summary>
    public float[]? Embedding { get; set; }

    /// <summary>
    /// Index position in the FAISS vector store (-1 if not added)
    /// </summary>
    public int VectorIndex { get; set; } = -1;
}
