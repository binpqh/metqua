using SimpleRag.Core;
using SimpleRag.Storage;

namespace SimpleRag.Services;

/// <summary>
/// Orchestrates the RAG pipeline: retrieval and generation
/// </summary>
public class RetrievalPipeline
{
    private readonly VectorStore _vectorStore;
    private readonly IModelClient _lmStudioClient;
    private readonly int _topK;

    public RetrievalPipeline(
        VectorStore vectorStore,
        IModelClient lmStudioClient,
        int topK = 5)
    {
        _vectorStore = vectorStore;
        _lmStudioClient = lmStudioClient;
        _topK = topK;
    }

    /// <summary>
    /// Retrieve relevant context and generate an answer
    /// </summary>
    public async Task<RagResponse> RetrieveAndAnswerAsync(
        string question,
        CancellationToken cancellationToken = default)
    {
        var startTime = DateTime.UtcNow;

        // Step 1: Embed the question
        Console.WriteLine("Embedding question...");
        var questionEmbedding = await _lmStudioClient.GetEmbeddingAsync(question, cancellationToken);

        // Step 2: Search for relevant chunks
        Console.WriteLine($"Searching for top {_topK} relevant chunks...");
        var searchResults = _vectorStore.SearchTopK(questionEmbedding, _topK);

        if (searchResults.Count == 0)
        {
            return new RagResponse
            {
                Question = question,
                Answer = "No relevant information found in the knowledge base.",
                RetrievedChunks = new List<RetrievedChunk>(),
                ProcessingTimeMs = (int)(DateTime.UtcNow - startTime).TotalMilliseconds
            };
        }

        // Step 3: Build context from retrieved chunks
        var context = BuildContext(searchResults);

        // Step 4: Generate answer using LLM
        Console.WriteLine("Generating answer...");
        var answer = await _lmStudioClient.GetCompletionAsync(question, context, cancellationToken);

        var response = new RagResponse
        {
            Question = question,
            Answer = answer,
            RetrievedChunks = searchResults.Select((sr, idx) => new RetrievedChunk
            {
                Text = sr.Chunk.Text,
                SourcePath = sr.Chunk.SourcePath,
                Score = sr.Score,
                Rank = idx + 1
            }).ToList(),
            ProcessingTimeMs = (int)(DateTime.UtcNow - startTime).TotalMilliseconds
        };

        return response;
    }

    /// <summary>
    /// Retrieve relevant chunks without generating an answer
    /// </summary>
    public async Task<List<SearchResult>> RetrieveAsync(
        string query,
        int? topK = null,
        CancellationToken cancellationToken = default)
    {
        var queryEmbedding = await _lmStudioClient.GetEmbeddingAsync(query, cancellationToken);
        return _vectorStore.SearchTopK(queryEmbedding, topK ?? _topK);
    }

    /// <summary>
    /// Build formatted context from search results
    /// </summary>
    private string BuildContext(List<SearchResult> searchResults)
    {
        var contextParts = searchResults.Select((result, index) =>
        {
            var sourceFile = Path.GetFileName(result.Chunk.SourcePath);
            return $"[Document {index + 1}] (Source: {sourceFile}, Relevance: {result.Score:F3})\n{result.Chunk.Text}";
        });

        return string.Join("\n\n---\n\n", contextParts);
    }
}

/// <summary>
/// Response from the RAG pipeline
/// </summary>
public class RagResponse
{
    public string Question { get; set; } = string.Empty;
    public string Answer { get; set; } = string.Empty;
    public List<RetrievedChunk> RetrievedChunks { get; set; } = new();
    public int ProcessingTimeMs { get; set; }

    public void PrintToConsole()
    {
        Console.WriteLine("\n" + new string('=', 80));
        Console.WriteLine("QUESTION:");
        Console.WriteLine(Question);
        Console.WriteLine("\n" + new string('-', 80));
        Console.WriteLine("ANSWER:");
        Console.WriteLine(Answer);
        Console.WriteLine("\n" + new string('-', 80));
        Console.WriteLine($"RETRIEVED CHUNKS ({RetrievedChunks.Count}):");

        foreach (var chunk in RetrievedChunks)
        {
            Console.WriteLine($"\n[{chunk.Rank}] {Path.GetFileName(chunk.SourcePath)} (Score: {chunk.Score:F3})");
            var preview = chunk.Text.Length > 150
                ? chunk.Text.Substring(0, 150) + "..."
                : chunk.Text;
            Console.WriteLine($"    {preview}");
        }

        Console.WriteLine("\n" + new string('-', 80));
        Console.WriteLine($"Processing time: {ProcessingTimeMs}ms");
        Console.WriteLine(new string('=', 80) + "\n");
    }
}

/// <summary>
/// A chunk retrieved from the vector store
/// </summary>
public class RetrievedChunk
{
    public string Text { get; set; } = string.Empty;
    public string SourcePath { get; set; } = string.Empty;
    public float Score { get; set; }
    public int Rank { get; set; }
}
