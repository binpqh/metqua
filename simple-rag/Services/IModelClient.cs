using System.Threading;
using System.Threading.Tasks;

namespace SimpleRag.Services;

/// <summary>
/// Abstract model client used by the application (embeddings + completion)
/// </summary>
public interface IModelClient : IDisposable
{
    Task<float[]> GetEmbeddingAsync(string text, CancellationToken cancellationToken = default);
    Task<List<float[]>> GetEmbeddingsBatchAsync(List<string> texts, CancellationToken cancellationToken = default);
    Task<string> GetCompletionAsync(string question, string context, CancellationToken cancellationToken = default);
    Task<bool> TestConnectionAsync();
}

