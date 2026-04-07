namespace CopiChat.Core.Models;

/// <summary>
/// GitHub Copilot authentication DTOs
/// </summary>
public sealed class DeviceCodeResponse
{
    public required string DeviceCode { get; init; }
    public required string UserCode { get; init; }
    public required string VerificationUri { get; init; }
    public required int ExpiresIn { get; init; }
    public required int Interval { get; init; }
}

public sealed class CopilotTokenResponse
{
    public required string Token { get; init; }
    public required long ExpiresAt { get; init; }
    public required string BaseUrl { get; init; }
}

/// <summary>
/// Chat DTOs
/// </summary>
public sealed class ChatMessage
{
    private static readonly string[] ValidRoles = ["system", "user", "assistant", "tool"];

    public required string Role { get; init; }
    public required string Content { get; init; }

    public bool IsValid() => ValidRoles.Contains(Role.ToLowerInvariant());
}

public sealed class ChatRequest
{
    public required string Model { get; init; }
    public required IReadOnlyList<ChatMessage> Messages { get; init; }
    public int? MaxTokens { get; init; }
    public double? Temperature { get; init; }
}

public sealed class ChatResponse
{
    public required string Id { get; init; }
    public required string Model { get; init; }
    public required IReadOnlyList<Choice> Choices { get; init; }
    public required Usage Usage { get; init; }
}

public sealed class Choice
{
    public required int Index { get; init; }
    public required ChatMessage Message { get; init; }
    public required string FinishReason { get; init; }
}

public sealed class Usage
{
    public required int PromptTokens { get; init; }
    public required int CompletionTokens { get; init; }
    public required int TotalTokens { get; init; }
}

/// <summary>
/// Model information
/// </summary>
public sealed class ModelInfo
{
    public required string Id { get; init; }
    public required string Name { get; init; }
    public required bool IsAvailable { get; init; }
}