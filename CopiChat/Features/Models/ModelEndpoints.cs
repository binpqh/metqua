using CopiChat.Core.Models;

namespace CopiChat.Features.Models;

/// <summary>
/// Model endpoints
/// </summary>
public static class ModelEndpoints
{
    public static RouteGroupBuilder MapModelEndpoints(this RouteGroupBuilder group)
    {
        group.MapGet("/", GetModels)
            .WithName("GetAvailableModels")
            .WithSummary("Get available Copilot models");

        return group;
    }

    private static async Task<IResult> GetModels(
        IModelService modelService,
        CancellationToken ct)
    {
        try
        {
            var models = await modelService.GetAvailableModelsAsync(ct);
            return Results.Ok(ApiResponse<IReadOnlyList<ModelInfo>>.Success(models));
        }
        catch (Exception ex)
        {
            return Results.Problem(ex.Message);
        }
    }
}