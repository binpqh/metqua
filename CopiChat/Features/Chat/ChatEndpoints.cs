using CopiChat.Core.Models;
using System.Text.Json;

namespace CopiChat.Features.Chat;

/// <summary>
/// Chat endpoints - Vertical Slice
/// </summary>
public static class ChatEndpoints
{
    public static RouteGroupBuilder MapChatEndpoints(this RouteGroupBuilder group)
    {
        group.MapPost("/completions", SendMessage)
            .WithName("SendChatMessage")
            .WithSummary("Send a chat message to GitHub Copilot");

        group.MapPost("/completions/stream", StreamMessage)
            .WithName("StreamChatMessage")
            .WithSummary("Stream chat message with tool calling support");

        return group;
    }

    private static async Task<IResult> SendMessage(
        ChatRequest request,
        IChatService chatService,
        CancellationToken ct)
    {
        try
        {
            var response = await chatService.SendMessageAsync(request, ct);
            return Results.Ok(ApiResponse<ChatResponse>.Success(response));
        }
        catch (ArgumentException ex)
        {
            return Results.BadRequest(ApiResponse<ChatResponse>.Failure(ex.Message));
        }
        catch (UnauthorizedAccessException)
        {
            return Results.Unauthorized();
        }
        catch (HttpRequestException ex)
        {
            return Results.BadRequest(ApiResponse<ChatResponse>.Failure($"API Error: {ex.Message}"));
        }
        catch (Exception ex)
        {
            return Results.Problem(ex.Message);
        }
    }

    private static async Task StreamMessage(
        HttpContext context,
        ChatWithToolsRequest request,
        IStreamingChatService streamingService)
    {
        context.Response.Headers.Append("Content-Type", "text/event-stream");
        context.Response.Headers.Append("Cache-Control", "no-cache");
        context.Response.Headers.Append("Connection", "keep-alive");

        try
        {
            await foreach (var evt in streamingService.StreamChatAsync(request, context.RequestAborted))
            {
                var eventType = evt.Type;
                var eventData = evt switch
                {
                    TextDeltaEvent text => (object)new { type = evt.Type, delta = text.Delta },
                    ToolCallStartEvent start => new { type = evt.Type, toolCall = new { start.ToolCall.Id, start.ToolCall.Name } },
                    ToolCallDeltaEvent delta => new { type = evt.Type, delta = delta.Delta },
                    ToolCallDoneEvent done => new { type = evt.Type, toolCall = new { done.ToolCall.Name, Arguments = done.ToolCall.Arguments.ToString() } },
                    DoneEvent => new { type = evt.Type },
                    _ => new { type = "unknown" }
                };

                await context.Response.WriteAsync($"event: {eventType}\n", context.RequestAborted);
                await context.Response.WriteAsync($"data: {JsonSerializer.Serialize(eventData)}\n\n", context.RequestAborted);
                await context.Response.Body.FlushAsync(context.RequestAborted);
            }
        }
        catch (Exception ex)
        {
            await context.Response.WriteAsync($"event: error\n", context.RequestAborted);
            await context.Response.WriteAsync($"data: {JsonSerializer.Serialize(new { error = ex.Message })}\n\n", context.RequestAborted);
        }
    }
}