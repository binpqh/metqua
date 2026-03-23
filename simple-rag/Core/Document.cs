namespace SimpleRag.Core;

/// <summary>
/// Represents a source document in the RAG system
/// </summary>
public class Document
{
    /// <summary>
    /// Unique identifier for the document
    /// </summary>
    public string Id { get; set; } = string.Empty;

    /// <summary>
    /// Full path to the document file
    /// </summary>
    public string FilePath { get; set; } = string.Empty;

    /// <summary>
    /// Text content of the document
    /// </summary>
    public string Content { get; set; } = string.Empty;

    /// <summary>
    /// When the document was added to the system
    /// </summary>
    public DateTime CreatedDate { get; set; } = DateTime.UtcNow;

    /// <summary>
    /// File size in bytes
    /// </summary>
    public long FileSize { get; set; }

    /// <summary>
    /// Number of chunks created from this document
    /// </summary>
    public int ChunkCount { get; set; }
}
