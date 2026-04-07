using CopiChat.Core.Models;
using Microsoft.AspNetCore.Mvc;

namespace CopiChat.Features.Auth;

/// <summary>
/// Authentication endpoints - Vertical Slice
/// Each endpoint is self-contained with minimal dependencies
/// </summary>
public static class AuthEndpoints
{
    public static RouteGroupBuilder MapAuthEndpoints(this RouteGroupBuilder group)
    {
        group.MapPost("/device-code", GetDeviceCode)
            .WithName("GetDeviceCode")
            .WithSummary("Initiate GitHub device flow authentication");

        group.MapPost("/poll", PollAccessToken)
            .WithName("PollAccessToken")
            .WithSummary("Poll for access token after user authentication");

        group.MapGet("/token", GetCopilotToken)
            .WithName("GetCopilotToken")
            .WithSummary("Get Copilot API token (auto-refreshed)");

        group.MapGet("/status", GetAuthStatus)
            .WithName("GetAuthStatus")
            .WithSummary("Check authentication status");

        group.MapPost("/logout", Logout)
            .WithName("Logout")
            .WithSummary("Clear all cached tokens");

        return group;
    }

    private static async Task<IResult> GetDeviceCode(
        IAuthService authService,
        CancellationToken ct)
    {
        try
        {
            var result = await authService.RequestDeviceCodeAsync(ct);
            return Results.Ok(ApiResponse<DeviceCodeResponse>.Success(
                result,
                $"Visit {result.VerificationUri} and enter code: {result.UserCode}"
            ));
        }
        catch (Exception ex)
        {
            return Results.BadRequest(ApiResponse<DeviceCodeResponse>.Failure(ex.Message));
        }
    }

    private static async Task<IResult> PollAccessToken(
        [FromQuery] string deviceCode,
        IAuthService authService,
        CancellationToken ct)
    {
        try
        {
            var token = await authService.PollAccessTokenAsync(deviceCode, ct);
            return Results.Ok(ApiResponse<string>.Success(token, "Authentication successful"));
        }
        catch (Exception ex)
        {
            return Results.BadRequest(ApiResponse<string>.Failure(ex.Message));
        }
    }

    private static async Task<IResult> GetCopilotToken(
        IAuthService authService,
        CancellationToken ct)
    {
        try
        {
            var token = await authService.GetCopilotTokenAsync(ct);
            return Results.Ok(ApiResponse<CopilotTokenResponse>.Success(token));
        }
        catch (UnauthorizedAccessException)
        {
            return Results.Unauthorized();
        }
        catch (Exception ex)
        {
            return Results.BadRequest(ApiResponse<CopilotTokenResponse>.Failure(ex.Message));
        }
    }

    private static IResult GetAuthStatus(IAuthService authService)
    {
        var isAuthenticated = authService.IsAuthenticated();
        return Results.Ok(ApiResponse<object>.Success(
            new { isAuthenticated },
            isAuthenticated ? "Authenticated" : "Not authenticated"
        ));
    }

    private static IResult Logout(IAuthService authService)
    {
        authService.Logout();
        return Results.Ok(ApiResponse<object>.Success(new { }, "Logged out successfully"));
    }
}