using System.Text.Json;

namespace CopiChat.Core.Models;

/// <summary>
/// Tool calling contracts
/// </summary>
public sealed class Tool
{
    public required string Type { get; init; } = "function";
    public required string Name { get; init; }
    public required string Description { get; init; }
    public required JsonElement Parameters { get; init; }
    public bool Strict { get; init; } = false;
}

public sealed class ToolCall
{
    public required string Id { get; init; }
    public required string CallId { get; init; }
    public required string Name { get; init; }
    public JsonElement Arguments { get; set; }
    public string PartialJson { get; set; } = string.Empty;
}

public sealed class ToolResult
{
    public required string CallId { get; init; }
    public required string Output { get; init; }
    public bool IsError { get; init; }
}

/// <summary>
/// Streaming event types
/// </summary>
public abstract class StreamEvent
{
    public required string Type { get; init; }
}

public sealed class TextDeltaEvent : StreamEvent
{
    public required string Delta { get; init; }
}

public sealed class ToolCallStartEvent : StreamEvent
{
    public required ToolCall ToolCall { get; init; }
}

public sealed class ToolCallDeltaEvent : StreamEvent
{
    public required ToolCall ToolCall { get; init; }
    public required string Delta { get; init; }
}

public sealed class ToolCallDoneEvent : StreamEvent
{
    public required ToolCall ToolCall { get; init; }
}

public sealed class DoneEvent : StreamEvent
{
}

/// <summary>
/// Enhanced chat request with tool support
/// </summary>
public sealed class ChatWithToolsRequest
{
    public required string Model { get; init; }
    public required IReadOnlyList<ChatMessage> Messages { get; init; }
    public IReadOnlyList<Tool>? Tools { get; init; }
    public int? MaxTokens { get; init; }
    public double? Temperature { get; init; }
    public bool Stream { get; init; }
}

