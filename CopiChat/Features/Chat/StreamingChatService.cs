using System.Runtime.CompilerServices;
using System.Text.Json;
using CopiChat.Core.Models;
using CopiChat.Features.Auth;
using CopiChat.Features.Tools;

namespace CopiChat.Features.Chat;

/// <summary>
/// Streaming chat service with tool calling support
/// </summary>
public interface IStreamingChatService
{
    IAsyncEnumerable<StreamEvent> StreamChatAsync(
        ChatWithToolsRequest request,
        CancellationToken ct = default);
}

public sealed class StreamingChatService : IStreamingChatService
{
    private readonly IAuthService _authService;
    private readonly IToolExecutor _toolExecutor;
    private static readonly HttpClient HttpClient = new();

    public StreamingChatService(IAuthService authService, IToolExecutor toolExecutor)
    {
        _authService = authService;
        _toolExecutor = toolExecutor;
    }

    public async IAsyncEnumerable<StreamEvent> StreamChatAsync(
        ChatWithToolsRequest request,
        [EnumeratorCancellation] CancellationToken ct = default)
    {
        var copilotToken = await _authService.GetCopilotTokenAsync(ct);
        var url = $"{copilotToken.BaseUrl}/chat/completions";

        var httpRequest = new HttpRequestMessage(HttpMethod.Post, url);
        AddHeaders(httpRequest, copilotToken.Token, request.Messages);

        var payload = BuildRequestPayload(request);
        httpRequest.Content = JsonContent.Create(payload);

        var response = await HttpClient.SendAsync(httpRequest, HttpCompletionOption.ResponseHeadersRead, ct);
        
        if (!response.IsSuccessStatusCode)
        {
            var error = await response.Content.ReadAsStringAsync(ct);
            throw new HttpRequestException($"API error: {error}");
        }

        var stream = await response.Content.ReadAsStreamAsync(ct);
        using var reader = new StreamReader(stream);

        ToolCall? currentToolCall = null;

        while (!reader.EndOfStream && !ct.IsCancellationRequested)
        {
            var line = await reader.ReadLineAsync(ct);
            if (string.IsNullOrWhiteSpace(line)) continue;

            if (!line.StartsWith("data: ")) continue;

            var data = line[6..]; // Remove "data: " prefix
            if (data == "[DONE]")
            {
                yield return new DoneEvent { Type = "done" };
                yield break;
            }

            JsonDocument? json;
            try
            {
                json = JsonDocument.Parse(data);
            }
            catch
            {
                continue; // Skip invalid JSON
            }

            var root = json.RootElement;
            
            // Handle tool calls
            if (root.TryGetProperty("choices", out var choices) && choices.GetArrayLength() > 0)
            {
                var choice = choices[0];
                
                if (choice.TryGetProperty("delta", out var delta))
                {
                    // Tool call start
                    if (delta.TryGetProperty("tool_calls", out var toolCalls) && toolCalls.GetArrayLength() > 0)
                    {
                        var toolCall = toolCalls[0];
                        
                        if (currentToolCall == null)
                        {
                            currentToolCall = new ToolCall
                            {
                                Id = toolCall.GetProperty("id").GetString() ?? "",
                                CallId = toolCall.GetProperty("id").GetString() ?? "",
                                Name = toolCall.TryGetProperty("function", out var func) 
                                    ? func.GetProperty("name").GetString() ?? ""
                                    : "",
                                PartialJson = ""
                            };

                            yield return new ToolCallStartEvent
                            {
                                Type = "tool_call_start",
                                ToolCall = currentToolCall
                            };
                        }

                        // Tool arguments delta
                        if (toolCall.TryGetProperty("function", out var function) &&
                            function.TryGetProperty("arguments", out var argsDelta))
                        {
                            var deltaStr = argsDelta.GetString() ?? "";
                            currentToolCall.PartialJson += deltaStr;

                            yield return new ToolCallDeltaEvent
                            {
                                Type = "tool_call_delta",
                                ToolCall = currentToolCall,
                                Delta = deltaStr
                            };
                        }
                    }
                    
                    // Text content delta
                    if (delta.TryGetProperty("content", out var contentDelta))
                    {
                        var text = contentDelta.GetString();
                        if (!string.IsNullOrEmpty(text))
                        {
                            yield return new TextDeltaEvent
                            {
                                Type = "text_delta",
                                Delta = text
                            };
                        }
                    }
                }

                // Check for finish_reason
                if (choice.TryGetProperty("finish_reason", out var finishReason) &&
                    finishReason.GetString() == "tool_calls" &&
                    currentToolCall != null)
                {
                    // Parse complete arguments
                    bool parseSuccess = false;
                    try
                    {
                        currentToolCall.Arguments = JsonDocument.Parse(currentToolCall.PartialJson).RootElement;
                        parseSuccess = true;
                    }
                    catch
                    {
                        // Failed to parse - skip tool execution
                        parseSuccess = false;
                    }

                    if (!parseSuccess)
                    {
                        yield return new DoneEvent { Type = "done" };
                        yield break;
                    }

                    yield return new ToolCallDoneEvent
                    {
                        Type = "tool_call_done",
                        ToolCall = currentToolCall
                    };

                    // Execute tool
                    var result = await _toolExecutor.ExecuteAsync(
                        currentToolCall.Name,
                        currentToolCall.Arguments,
                        ct);

                    // Continue conversation with tool result
                    var newMessages = request.Messages.ToList();
                    newMessages.Add(new ChatMessage
                    {
                        Role = "assistant",
                        Content = $"[Tool Call: {currentToolCall.Name}]"
                    });
                    newMessages.Add(new ChatMessage
                    {
                        Role = "tool",
                        Content = result.Output
                    });

                    var newRequest = new ChatWithToolsRequest
                    {
                        Model = request.Model,
                        Messages = newMessages,
                        Tools = request.Tools,
                        MaxTokens = request.MaxTokens,
                        Temperature = request.Temperature,
                        Stream = true
                    };

                    // Recursive call with tool result
                    await foreach (var evt in StreamChatAsync(newRequest, ct))
                    {
                        yield return evt;
                    }

                    yield break;
                }
            }
        }
    }

    private static void AddHeaders(HttpRequestMessage request, string token, IReadOnlyList<ChatMessage> messages)
    {
        request.Headers.Add("Authorization", $"Bearer {token}");
        request.Headers.Add("Accept", "text/event-stream");
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

    private static object BuildRequestPayload(ChatWithToolsRequest request)
    {
        var payload = new Dictionary<string, object>
        {
            ["model"] = request.Model,
            ["messages"] = request.Messages.Select(m => new
            {
                role = m.Role.ToLowerInvariant(),
                content = m.Content
            }).ToArray(),
            ["max_tokens"] = request.MaxTokens ?? 4096,
            ["temperature"] = request.Temperature ?? 0.7,
            ["stream"] = true
        };

        if (request.Tools?.Count > 0)
        {
            payload["tools"] = request.Tools.Select(t => new
            {
                type = t.Type,
                function = new
                {
                    name = t.Name,
                    description = t.Description,
                    parameters = JsonSerializer.Serialize(t.Parameters)
                }
            }).ToArray();
        }

        return payload;
    }
}



