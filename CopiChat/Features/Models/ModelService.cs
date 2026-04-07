using CopiChat.Core.Models;

namespace CopiChat.Features.Models;

/// <summary>
/// Model information service
/// </summary>
public interface IModelService
{
    Task<IReadOnlyList<ModelInfo>> GetAvailableModelsAsync(CancellationToken ct = default);
}

public sealed class ModelService : IModelService
{
    private static readonly ModelInfo[] DefaultModels =
    [
        new() { Id = "gpt-4o-2024-08-06", Name = "GPT-4o", IsAvailable = true },
        new() { Id = "gpt-4o-mini-2024-07-18", Name = "GPT-4o Mini", IsAvailable = true },
        new() { Id = "claude-3-5-sonnet-20241022", Name = "Claude 3.5 Sonnet", IsAvailable = true },
        new() { Id = "o1-preview-2024-09-12", Name = "o1 Preview", IsAvailable = false },
        new() { Id = "o1-mini-2024-09-12", Name = "o1 Mini", IsAvailable = false }
    ];

    public Task<IReadOnlyList<ModelInfo>> GetAvailableModelsAsync(CancellationToken ct = default)
    {
        // Future: Query GitHub API for a real model list
        return Task.FromResult<IReadOnlyList<ModelInfo>>(DefaultModels);
    }
}