using System.Net.Http.Json;
using System.Text.Json;
using System.Text.Json.Serialization;
using SimpleRag.Core;

namespace SimpleRag.Services;

/// <summary>
/// Client for interacting with LMStudio's local API (OpenAI-compatible)
/// </summary>
public class LMStudioClient : IModelClient
{
    private readonly HttpClient _httpClient;
    private readonly LMStudioConfig _config;
    private readonly JsonSerializerOptions _jsonOptions;

    public LMStudioClient(LMStudioConfig config)
    {
        _config = config;
        _httpClient = new HttpClient
        {
            BaseAddress = new Uri(config.BaseUrl),
            Timeout = TimeSpan.FromSeconds(config.Timeout)
        };

        _jsonOptions = new JsonSerializerOptions
        {
            PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
            DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull
        };
    }

    /// <summary>
    /// Get embedding vector for a single text
    /// </summary>
    public async Task<float[]> GetEmbeddingAsync(string text, CancellationToken cancellationToken = default)
    {
        var request = new EmbeddingRequest
        {
            Input = text,
            Model = _config.EmbeddingModel
        };

        try
        {
            var response = await _httpClient.PostAsJsonAsync("/v1/embeddings", request, _jsonOptions, cancellationToken);
            response.EnsureSuccessStatusCode();

            var result = await response.Content.ReadFromJsonAsync<EmbeddingResponse>(_jsonOptions, cancellationToken);

            if (result?.Data == null || result.Data.Count == 0)
            {
                throw new InvalidOperationException("No embedding data returned from LMStudio");
            }

            return result.Data[0].Embedding;
        }
        catch (HttpRequestException ex)
        {
            throw new InvalidOperationException(
                $"Failed to get embedding from LMStudio at {_config.BaseUrl}. " +
                "Ensure LMStudio is running and the embedding model is loaded.", ex);
        }
    }

    /// <summary>
    /// Get embeddings for multiple texts in a single batch
    /// </summary>
    public async Task<List<float[]>> GetEmbeddingsBatchAsync(List<string> texts, CancellationToken cancellationToken = default)
    {
        var request = new EmbeddingRequest
        {
            Input = texts,
            Model = _config.EmbeddingModel
        };

        try
        {
            var response = await _httpClient.PostAsJsonAsync("/v1/embeddings", request, _jsonOptions, cancellationToken);
            response.EnsureSuccessStatusCode();

            var result = await response.Content.ReadFromJsonAsync<EmbeddingResponse>(_jsonOptions, cancellationToken);

            if (result?.Data == null)
            {
                throw new InvalidOperationException("No embedding data returned from LMStudio");
            }

            return result.Data.Select(d => d.Embedding).ToList();
        }
        catch (HttpRequestException ex)
        {
            throw new InvalidOperationException(
                $"Failed to get embeddings from LMStudio at {_config.BaseUrl}. " +
                "Ensure LMStudio is running and the embedding model is loaded.", ex);
        }
    }

    /// <summary>
    /// Get completion from LLM with context
    /// </summary>
    public async Task<string> GetCompletionAsync(string question, string context, CancellationToken cancellationToken = default)
    {
        var systemPrompt = "You are a helpful AI assistant. Answer questions based on the provided context. " +
                          "If the context doesn't contain relevant information, say so clearly.";

        var userPrompt = $@"Context from knowledge base:
---
{context}
---

Question: {question}

Answer the question based on the context above. Be specific and cite relevant information.";

        var request = new CompletionRequest
        {
            Model = _config.CompletionModel,
            Messages = new[]
            {
                new Message { Role = "system", Content = systemPrompt },
                new Message { Role = "user", Content = userPrompt }
            },
            Temperature = 0.7,
            MaxTokens = 512
        };

        try
        {
            var response = await _httpClient.PostAsJsonAsync("/v1/chat/completions", request, _jsonOptions, cancellationToken);
            response.EnsureSuccessStatusCode();

            var result = await response.Content.ReadFromJsonAsync<CompletionResponse>(_jsonOptions, cancellationToken);

            if (result?.Choices == null || result.Choices.Count == 0)
            {
                throw new InvalidOperationException("No completion returned from LMStudio");
            }

            return result.Choices[0].Message.Content;
        }
        catch (HttpRequestException ex)
        {
            throw new InvalidOperationException(
                $"Failed to get completion from LMStudio at {_config.BaseUrl}. " +
                "Ensure LMStudio is running and a language model is loaded.", ex);
        }
    }

    /// <summary>
    /// Test connection to LMStudio
    /// </summary>
    public async Task<bool> TestConnectionAsync()
    {
        try
        {
            var response = await _httpClient.GetAsync("/v1/models");
            return response.IsSuccessStatusCode;
        }
        catch
        {
            return false;
        }
    }

    public void Dispose()
    {
        _httpClient?.Dispose();
    }

    #region Request/Response Models

    private class EmbeddingRequest
    {
        [JsonPropertyName("input")]
        public object Input { get; set; } = ""; // Can be string or List<string>

        [JsonPropertyName("model")]
        public string Model { get; set; } = "";
    }

    private class EmbeddingResponse
    {
        [JsonPropertyName("data")]
        public List<EmbeddingData> Data { get; set; } = new();
    }

    private class EmbeddingData
    {
        [JsonPropertyName("embedding")]
        public float[] Embedding { get; set; } = Array.Empty<float>();

        [JsonPropertyName("index")]
        public int Index { get; set; }
    }

    private class CompletionRequest
    {
        [JsonPropertyName("model")]
        public string Model { get; set; } = "";

        [JsonPropertyName("messages")]
        public Message[] Messages { get; set; } = Array.Empty<Message>();

        [JsonPropertyName("temperature")]
        public double Temperature { get; set; } = 0.7;

        [JsonPropertyName("max_tokens")]
        public int MaxTokens { get; set; } = 512;
    }

    private class Message
    {
        [JsonPropertyName("role")]
        public string Role { get; set; } = "";

        [JsonPropertyName("content")]
        public string Content { get; set; } = "";
    }

    private class CompletionResponse
    {
        [JsonPropertyName("choices")]
        public List<Choice> Choices { get; set; } = new();
    }

    private class Choice
    {
        [JsonPropertyName("message")]
        public Message Message { get; set; } = new();

        [JsonPropertyName("finish_reason")]
        public string FinishReason { get; set; } = "";
    }

    #endregion
}
