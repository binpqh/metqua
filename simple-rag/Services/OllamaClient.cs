using System.Net.Http.Json;
using System.Text.Json;
using System.Text.Json.Serialization;
using SimpleRag.Core;

namespace SimpleRag.Services;

/// <summary>
/// Client for interacting with Ollama local server
/// </summary>
public class OllamaClient : IModelClient
{
    private readonly HttpClient _httpClient;
    private readonly OllamaConfig _config;
    private readonly JsonSerializerOptions _jsonOptions;

    public OllamaClient(OllamaConfig config)
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

    public async Task<float[]> GetEmbeddingAsync(string text, CancellationToken cancellationToken = default)
    {
        var request = new { model = _config.EmbeddingModel, prompt = text };
        try
        {
            var response = await _httpClient.PostAsJsonAsync($"/embeddings/", request, _jsonOptions, cancellationToken);
            response.EnsureSuccessStatusCode();

            var resp = await response.Content.ReadFromJsonAsync<OllamaEmbedResponse>(_jsonOptions, cancellationToken);
            if (resp?.Embedding == null)
                throw new InvalidOperationException("No embedding returned from Ollama");
            return resp.Embedding;
        }
        catch (HttpRequestException ex)
        {
            throw new InvalidOperationException($"Failed to get embedding from Ollama at {_config.BaseUrl}", ex);
        }
    }

    public async Task<List<float[]>> GetEmbeddingsBatchAsync(List<string> texts, CancellationToken cancellationToken = default)
    {
        var results = new List<float[]>();
        foreach (var t in texts)
        {
            var emb = await GetEmbeddingAsync(t, cancellationToken);
            results.Add(emb);
        }
        return results;
    }

    public async Task<string> GetCompletionAsync(string question, string context, CancellationToken cancellationToken = default)
    {
        var prompt = $"Context:\n{context}\n\nQuestion: {question}\n\nAnswer:";

        var request = new { model = _config.CompletionModel, prompt = prompt, max_tokens = 512, temperature = 0.7 };

        try
        {
            var response = await _httpClient.PostAsJsonAsync($"/chat/{_config.CompletionModel}", request, _jsonOptions, cancellationToken);
            response.EnsureSuccessStatusCode();

            var resp = await response.Content.ReadFromJsonAsync<OllamaChatResponse>(_jsonOptions, cancellationToken);
            if (resp?.Text == null)
                throw new InvalidOperationException("No completion returned from Ollama");
            return resp.Text;
        }
        catch (HttpRequestException ex)
        {
            throw new InvalidOperationException($"Failed to get completion from Ollama at {_config.BaseUrl}", ex);
        }
    }

    public async Task<bool> TestConnectionAsync()
    {
        // Try several known model-list endpoints that different Ollama versions expose
        var probes = new[] { "/models", "/api/models", "/v1/models" };

        foreach (var p in probes)
        {
            try
            {
                var resp = await _httpClient.GetAsync(p);
                if (resp.IsSuccessStatusCode)
                {
                    Console.WriteLine($"  ↳ probe {p} succeeded ({(int)resp.StatusCode})");
                    return true;
                }

                var text = await resp.Content.ReadAsStringAsync();
                Console.WriteLine($"  ↳ probe {p} returned {(int)resp.StatusCode}: {Truncate(text)}");
            }
            catch (Exception ex)
            {
                Console.WriteLine($"  ↳ probe {p} failed: {ex.Message}");
            }
        }

        // Try embedding endpoints as a fallback
        var embedProbes = new[] { $"/embed/{_config.EmbeddingModel}", "/embed", "/embeddings", $"/embeddings/{_config.EmbeddingModel}" };
        var pingReq = new { model = _config.EmbeddingModel, input = "ping" };

        foreach (var ep in embedProbes)
        {
            try
            {
                var resp = await _httpClient.PostAsJsonAsync(ep, pingReq, _jsonOptions);
                if (resp.IsSuccessStatusCode)
                {
                    Console.WriteLine($"  ↳ embed probe {ep} succeeded ({(int)resp.StatusCode})");
                    try
                    {
                        var body = await resp.Content.ReadFromJsonAsync<OllamaEmbedResponse>(_jsonOptions);
                        if (body?.Embedding != null && body.Embedding.Length > 0)
                            return true;
                    }
                    catch (Exception ex)
                    {
                        Console.WriteLine($"    (parsing embedding response failed: {ex.Message})");
                        return true; // server responded OK
                    }
                }

                var text = await resp.Content.ReadAsStringAsync();
                Console.WriteLine($"  ↳ embed probe {ep} returned {(int)resp.StatusCode}: {Truncate(text)}");
            }
            catch (Exception ex)
            {
                Console.WriteLine($"  ↳ embed probe {ep} failed: {ex.Message}");
            }
        }

        // Try chat/completion endpoints as another fallback
        var chatProbes = new[] { $"/chat/{_config.CompletionModel}", "/chat", "/completions", "/v1/chat/completions" };
        var chatReq = new { model = _config.CompletionModel, prompt = "ping", max_tokens = 1 };

        foreach (var ep in chatProbes)
        {
            try
            {
                var resp = await _httpClient.PostAsJsonAsync(ep, chatReq, _jsonOptions);
                if (resp.IsSuccessStatusCode)
                {
                    Console.WriteLine($"  ↳ chat probe {ep} succeeded ({(int)resp.StatusCode})");
                    return true;
                }

                var text = await resp.Content.ReadAsStringAsync();
                Console.WriteLine($"  ↳ chat probe {ep} returned {(int)resp.StatusCode}: {Truncate(text)}");
            }
            catch (Exception ex)
            {
                Console.WriteLine($"  ↳ chat probe {ep} failed: {ex.Message}");
            }
        }

        return false;
    }

    private static string Truncate(string s, int max = 200)
    {
        if (string.IsNullOrEmpty(s)) return string.Empty;
        return s.Length <= max ? s : s.Substring(0, max) + "...";
    }

    public void Dispose()
    {
        _httpClient?.Dispose();
    }

    private class OllamaEmbedResponse
    {
        [JsonPropertyName("embedding")]
        public float[] Embedding { get; set; } = Array.Empty<float>();
    }

    private class OllamaChatResponse
    {
        [JsonPropertyName("text")]
        public string Text { get; set; } = string.Empty;
    }
}

