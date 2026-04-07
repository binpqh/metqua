using CopiChat.Core.Models;
using CopiChat.Features.Auth;

namespace CopiChat.Features.Chat;

/// <summary>
/// GitHub Copilot Chat service
/// Uses /chat/completions endpoint (OpenAI-compatible)
/// </summary>
public interface IChatService
{
    Task<ChatResponse> SendMessageAsync(ChatRequest request, CancellationToken ct = default);
}

public sealed class ChatService : IChatService
{
    private readonly IAuthService _authService;
    private static readonly HttpClient HttpClient = new();

    public ChatService(IAuthService authService)
    {
        _authService = authService;
    }

    public async Task<ChatResponse> SendMessageAsync(ChatRequest request, CancellationToken ct = default)
    {
        ValidateRequest(request);

        var copilotToken = await _authService.GetCopilotTokenAsync(ct);
        var url = $"{copilotToken.BaseUrl}/chat/completions";

        var httpRequest = new HttpRequestMessage(HttpMethod.Post, url);
        AddHeaders(httpRequest, copilotToken.Token, request.Messages);

        var payload = new
        {
            model = request.Model,
            messages = request.Messages.Select(m => new
            {
                role = m.Role.ToLowerInvariant(),
                content = m.Content
            }).ToArray(),
            max_tokens = request.MaxTokens ?? 4096,
            temperature = request.Temperature ?? 0.7,
            stream = false
        };

        httpRequest.Content = JsonContent.Create(payload);
        var response = await HttpClient.SendAsync(httpRequest, ct);

        if (!response.IsSuccessStatusCode)
        {
            var errorBody = await response.Content.ReadAsStringAsync(ct);
            throw new HttpRequestException($"API error ({response.StatusCode}): {errorBody}");
        }

        var result = await response.Content.ReadFromJsonAsync<ApiResponseDto>(ct);
        return MapToResponse(result!);
    }

    private static void AddHeaders(HttpRequestMessage request, string token, IReadOnlyList<ChatMessage> messages)
    {
        request.Headers.Add("Authorization", $"Bearer {token}");
        request.Headers.Add("Accept", "application/json");
        request.Headers.Add("User-Agent", "GitHubCopilotChat/0.35.0");
        request.Headers.Add("Editor-Version", "vscode/1.107.0");
        request.Headers.Add("Editor-Plugin-Version", "copilot-chat/0.35.0");
        request.Headers.Add("Copilot-Integration-Id", "vscode-chat");
        request.Headers.Add("X-Github-Api-Version", "2023-07-07");
        request.Headers.Add("Openai-Intent", "conversation-panel");

        var lastMessage = messages.LastOrDefault();
        var initiator = lastMessage?.Role?.ToLowerInvariant() == "user" ? "user" : "agent";
        request.Headers.Add("X-Initiator", initiator);
    }

    private static void ValidateRequest(ChatRequest request)
    {
        if (string.IsNullOrWhiteSpace(request.Model))
            throw new ArgumentException("Model is required");

        if (request.Messages?.Count == 0)
            throw new ArgumentException("Messages are required");

        foreach (var msg in request.Messages!)
        {
            if (!msg.IsValid())
                throw new ArgumentException($"Invalid role: {msg.Role}");
        }
    }

    private static ChatResponse MapToResponse(ApiResponseDto dto)
    {
        var firstChoice = dto.choices?.FirstOrDefault();

        return new ChatResponse
        {
            Id = dto.id ?? "unknown",
            Model = dto.model ?? "unknown",
            Choices = new List<Choice>
            {
                new Choice
                {
                    Index = firstChoice?.index ?? 0,
                    Message = new ChatMessage
                    {
                        Role = firstChoice?.message?.role ?? "assistant",
                        Content = firstChoice?.message?.content ?? string.Empty
                    },
                    FinishReason = firstChoice?.finish_reason ?? "stop"
                }
            },
            Usage = new Usage
            {
                PromptTokens = dto.usage?.prompt_tokens ?? 0,
                CompletionTokens = dto.usage?.completion_tokens ?? 0,
                TotalTokens = dto.usage?.total_tokens ?? 0
            }
        };
    }

    // Internal DTOs
    private sealed record ApiResponseDto(
        string? id,
        string? model,
        List<ChoiceDto>? choices,
        UsageDto? usage
    );

    private sealed record ChoiceDto(
        int index,
        MessageDto? message,
        string? finish_reason
    );

    private sealed record MessageDto(
        string? role,
        string? content
    );

    private sealed record UsageDto(
        int prompt_tokens,
        int completion_tokens,
        int total_tokens
    );
}