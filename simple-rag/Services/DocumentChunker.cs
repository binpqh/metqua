using SimpleRag.Core;

namespace SimpleRag.Services;

/// <summary>
/// Splits documents into overlapping chunks for embedding
/// </summary>
public class DocumentChunker
{
    private readonly int _chunkSize;
    private readonly int _chunkOverlap;

    public DocumentChunker(int chunkSize = 512, int chunkOverlap = 64)
    {
        if (chunkSize <= 0)
            throw new ArgumentException("Chunk size must be positive", nameof(chunkSize));

        if (chunkOverlap < 0 || chunkOverlap >= chunkSize)
            throw new ArgumentException("Overlap must be non-negative and less than chunk size", nameof(chunkOverlap));

        _chunkSize = chunkSize;
        _chunkOverlap = chunkOverlap;
    }

    /// <summary>
    /// Chunk a document into overlapping text segments
    /// </summary>
    public List<Chunk> ChunkDocument(Document document)
    {
        var chunks = new List<Chunk>();

        if (string.IsNullOrWhiteSpace(document.Content))
        {
            return chunks;
        }

        // Split by whitespace to get words
        var words = document.Content
            .Split(new[] { ' ', '\t', '\n', '\r' }, StringSplitOptions.RemoveEmptyEntries)
            .ToList();

        if (words.Count == 0)
        {
            return chunks;
        }

        int chunkIndex = 0;
        int position = 0;

        // Create overlapping chunks
        while (position < words.Count)
        {
            // Take up to chunkSize words
            var chunkWords = words
                .Skip(position)
                .Take(_chunkSize)
                .ToList();

            if (chunkWords.Count == 0)
                break;

            var chunkText = string.Join(" ", chunkWords);

            // Clean up excessive whitespace
            chunkText = System.Text.RegularExpressions.Regex.Replace(chunkText, @"\s+", " ").Trim();

            var chunk = new Chunk
            {
                Id = $"{document.Id}_chunk_{chunkIndex}",
                Text = chunkText,
                Tokens = chunkWords.Count,
                SourceId = document.Id,
                SourcePath = document.FilePath,
                Position = position
            };

            chunks.Add(chunk);
            chunkIndex++;

            // Move forward by (chunkSize - overlap) to create overlap
            position += _chunkSize - _chunkOverlap;

            // If we're at or past the end, break
            if (position >= words.Count)
                break;
        }

        return chunks;
    }

    /// <summary>
    /// Chunk multiple documents
    /// </summary>
    public Dictionary<string, List<Chunk>> ChunkDocuments(List<Document> documents)
    {
        var result = new Dictionary<string, List<Chunk>>();

        foreach (var document in documents)
        {
            var chunks = ChunkDocument(document);
            result[document.Id] = chunks;
        }

        return result;
    }

    /// <summary>
    /// Get statistics about chunking configuration
    /// </summary>
    public string GetStats()
    {
        return $"Chunk Size: {_chunkSize} words\n" +
               $"Overlap: {_chunkOverlap} words\n" +
               $"Effective Step: {_chunkSize - _chunkOverlap} words";
    }
}
