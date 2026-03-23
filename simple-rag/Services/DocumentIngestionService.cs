using SimpleRag.Core;
using SimpleRag.Storage;

namespace SimpleRag.Services;

/// <summary>
/// Handles document ingestion into the RAG system
/// </summary>
public class DocumentIngestionService
{
    private readonly DocumentChunker _chunker;
    private readonly IModelClient _lmStudioClient;
    private readonly VectorStore _vectorStore;

    public DocumentIngestionService(
        DocumentChunker chunker,
        IModelClient lmStudioClient,
        VectorStore vectorStore)
    {
        _chunker = chunker;
        _lmStudioClient = lmStudioClient;
        _vectorStore = vectorStore;
    }

    /// <summary>
    /// Ingest all documents from a directory
    /// </summary>
    public async Task<IngestionResult> IngestDirectoryAsync(
        string directoryPath,
        CancellationToken cancellationToken = default)
    {
        var result = new IngestionResult { StartTime = DateTime.UtcNow };

        if (!Directory.Exists(directoryPath))
        {
            result.Errors.Add($"Directory not found: {directoryPath}");
            return result;
        }

        // Find all text and markdown files
        var supportedExtensions = new[] { ".txt", ".md" };
        var files = Directory.GetFiles(directoryPath, "*.*", SearchOption.AllDirectories)
            .Where(f => supportedExtensions.Contains(Path.GetExtension(f).ToLowerInvariant()))
            .ToList();

        if (files.Count == 0)
        {
            result.Errors.Add($"No .txt or .md files found in {directoryPath}");
            return result;
        }

        Console.WriteLine($"Found {files.Count} documents to ingest...\n");

        foreach (var filePath in files)
        {
            try
            {
                await IngestFileAsync(filePath, result, cancellationToken);
            }
            catch (Exception ex)
            {
                result.Errors.Add($"Failed to ingest {filePath}: {ex.Message}");
            }
        }

        result.EndTime = DateTime.UtcNow;
        return result;
    }

    /// <summary>
    /// Ingest a single file
    /// </summary>
    public async Task IngestFileAsync(
        string filePath,
        IngestionResult result,
        CancellationToken cancellationToken = default)
    {
        var fileName = Path.GetFileName(filePath);
        Console.WriteLine($"Processing: {fileName}");

        // Read file content
        var content = await File.ReadAllTextAsync(filePath, cancellationToken);
        var fileInfo = new FileInfo(filePath);

        // Create document
        var document = new Document
        {
            Id = Guid.NewGuid().ToString(),
            FilePath = filePath,
            Content = content,
            FileSize = fileInfo.Length,
            CreatedDate = DateTime.UtcNow
        };

        // Chunk the document
        var chunks = _chunker.ChunkDocument(document);
        document.ChunkCount = chunks.Count;

        Console.WriteLine($"  - Created {chunks.Count} chunks");

        if (chunks.Count == 0)
        {
            result.Errors.Add($"No chunks created from {fileName}");
            return;
        }

        // Embed chunks in batches
        const int batchSize = 10; // Process 10 chunks at a time
        for (int i = 0; i < chunks.Count; i += batchSize)
        {
            var batch = chunks.Skip(i).Take(batchSize).ToList();
            var texts = batch.Select(c => c.Text).ToList();

            Console.WriteLine($"  - Embedding chunks {i + 1}-{Math.Min(i + batchSize, chunks.Count)}...");

            try
            {
                // Get embeddings for the batch
                var embeddings = await _lmStudioClient.GetEmbeddingsBatchAsync(texts, cancellationToken);

                // Assign embeddings to chunks
                for (int j = 0; j < batch.Count; j++)
                {
                    batch[j].Embedding = embeddings[j];
                    _vectorStore.AddChunk(batch[j]);
                }

                result.ChunksProcessed += batch.Count;
            }
            catch (Exception ex)
            {
                result.Errors.Add($"Failed to embed chunks from {fileName}: {ex.Message}");
                throw;
            }
        }

        result.FilesProcessed++;
        result.Documents.Add(document);
        Console.WriteLine($"  ✓ Completed {fileName}\n");
    }

    /// <summary>
    /// Get current ingestion statistics
    /// </summary>
    public IngestionStats GetStats()
    {
        var storeStats = _vectorStore.GetStats();

        return new IngestionStats
        {
            TotalChunks = storeStats.TotalChunks,
            TotalDocuments = storeStats.UniqueDocuments,
            VectorDimension = storeStats.Dimension
        };
    }
}

/// <summary>
/// Result of a document ingestion operation
/// </summary>
public class IngestionResult
{
    public DateTime StartTime { get; set; }
    public DateTime EndTime { get; set; }
    public int FilesProcessed { get; set; }
    public int ChunksProcessed { get; set; }
    public List<Document> Documents { get; set; } = new();
    public List<string> Errors { get; set; } = new();

    public TimeSpan Duration => EndTime - StartTime;

    public void PrintSummary()
    {
        Console.WriteLine("\n" + new string('=', 80));
        Console.WriteLine("INGESTION SUMMARY");
        Console.WriteLine(new string('=', 80));
        Console.WriteLine($"Files processed: {FilesProcessed}");
        Console.WriteLine($"Chunks created: {ChunksProcessed}");
        Console.WriteLine($"Duration: {Duration.TotalSeconds:F2} seconds");

        if (Errors.Count > 0)
        {
            Console.WriteLine($"\nErrors ({Errors.Count}):");
            foreach (var error in Errors)
            {
                Console.WriteLine($"  - {error}");
            }
        }

        Console.WriteLine(new string('=', 80) + "\n");
    }
}

/// <summary>
/// Statistics about ingested data
/// </summary>
public class IngestionStats
{
    public int TotalDocuments { get; set; }
    public int TotalChunks { get; set; }
    public int VectorDimension { get; set; }
}
