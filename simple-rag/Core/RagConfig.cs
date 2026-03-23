namespace SimpleRag.Core;

/// <summary>
/// Configuration settings for the RAG system
/// </summary>
public class RagConfig
{
    // Backwards-compatible LMStudio config
    public LMStudioConfig LMStudio { get; set; } = new();

    // Which provider to use: "lmstudio" (default) or "ollama"
    public string Provider { get; set; } = "lmstudio";

    // Ollama configuration (used when Provider == "ollama")
    public OllamaConfig Ollama { get; set; } = new();
    public RAGSettings RAG { get; set; } = new();
}

public class LMStudioConfig
{
    public string BaseUrl { get; set; } = "http://localhost:1234";
    public string EmbeddingModel { get; set; } = "nomic-embed-text";
    public string CompletionModel { get; set; } = "llama-7b";
    public int EmbeddingDimension { get; set; } = 768;
    public int Timeout { get; set; } = 60;
}

public class OllamaConfig
{
    public string BaseUrl { get; set; } = "http://localhost:11434/api"; // Ollama default
    public string EmbeddingModel { get; set; } = "qwen3-embedding:0.6b";
    public string CompletionModel { get; set; } = "qwen3.5:0.8b";
    public int EmbeddingDimension { get; set; } = 768;
    public int Timeout { get; set; } = 60;
}

public class RAGSettings
{
    public int MaxChunksPerQuery { get; set; } = 5;
    public int ChunkSize { get; set; } = 512;
    public int ChunkOverlap { get; set; } = 64;
    public int MaxDocuments { get; set; } = 100;
    public string DocumentsPath { get; set; } = "./documents";
}
