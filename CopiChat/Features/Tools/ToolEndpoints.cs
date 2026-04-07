using CopiChat.Core.Models;

namespace CopiChat.Features.Tools;

/// <summary>
/// Tool endpoints
/// </summary>
public static class ToolEndpoints
{
    public static RouteGroupBuilder MapToolEndpoints(this RouteGroupBuilder group)
    {
        group.MapGet("/", GetTools)
            .WithName("GetAvailableTools")
            .WithSummary("Get available tools for Copilot");

        return group;
    }

    private static IResult GetTools(IToolService toolService)
    {
        try
        {
            var tools = toolService.GetAvailableTools();
            return Results.Ok(ApiResponse<IReadOnlyList<Tool>>.Success(
                tools,
                $"Available tools: {tools.Count}"
            ));
        }
        catch (Exception ex)
        {
            return Results.Problem(ex.Message);
        }
    }
}

