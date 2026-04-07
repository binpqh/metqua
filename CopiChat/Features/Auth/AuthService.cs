using CopiChat.Core.Caching;
using CopiChat.Core.Models;

namespace CopiChat.Features.Auth;

/// <summary>
/// GitHub Copilot authentication service
/// Vertical Slice: Contains all auth-related logic
/// </summary>
public interface IAuthService
{
    Task<DeviceCodeResponse> RequestDeviceCodeAsync(CancellationToken ct = default);
    Task<string> PollAccessTokenAsync(string deviceCode, CancellationToken ct = default);
    Task<CopilotTokenResponse> GetCopilotTokenAsync(CancellationToken ct = default);
    bool IsAuthenticated();
    void Logout();
}

public sealed class AuthService : IAuthService
{
    private const string ClientId = "Iv1.b507a08c87ecfe98";
    private const string DeviceCodeUrl = "https://github.com/login/device/code";
    private const string AccessTokenUrl = "https://github.com/login/oauth/access_token";
    private const string CopilotTokenUrl = "https://api.github.com/copilot_internal/v2/token";

    private static readonly HttpClient HttpClient = new();
    private readonly ICache<string> _accessTokenCache;
    private readonly ITokenCache _tokenCache;

    public AuthService(ICache<string> accessTokenCache, ITokenCache tokenCache)
    {
        _accessTokenCache = accessTokenCache;
        _tokenCache = tokenCache;
    }

    public async Task<DeviceCodeResponse> RequestDeviceCodeAsync(CancellationToken ct = default)
    {
        var formData = new Dictionary<string, string>
        {
            ["client_id"] = ClientId,
            ["scope"] = "read:user"
        };

        using var content = new FormUrlEncodedContent(formData);
        using var request = new HttpRequestMessage(HttpMethod.Post, DeviceCodeUrl)
        {
            Content = content
        };
        request.Headers.Add("Accept", "application/json");

        var response = await HttpClient.SendAsync(request, ct);
        response.EnsureSuccessStatusCode();

        var result = await response.Content.ReadFromJsonAsync<DeviceCodeDto>(ct);
        return new DeviceCodeResponse
        {
            DeviceCode = result!.device_code,
            UserCode = result.user_code,
            VerificationUri = result.verification_uri,
            ExpiresIn = result.expires_in,
            Interval = result.interval
        };
    }

    public async Task<string> PollAccessTokenAsync(string deviceCode, CancellationToken ct = default)
    {
        var formData = new Dictionary<string, string>
        {
            ["client_id"] = ClientId,
            ["device_code"] = deviceCode,
            ["grant_type"] = "urn:ietf:params:oauth:grant-type:device_code"
        };

        var expiresAt = DateTime.UtcNow.AddMinutes(15);

        while (DateTime.UtcNow < expiresAt && !ct.IsCancellationRequested)
        {
            using var content = new FormUrlEncodedContent(formData);
            using var request = new HttpRequestMessage(HttpMethod.Post, AccessTokenUrl)
            {
                Content = content
            };
            request.Headers.Add("Accept", "application/json");

            var response = await HttpClient.SendAsync(request, ct);
            var result = await response.Content.ReadFromJsonAsync<AccessTokenDto>(ct);

            if (!string.IsNullOrEmpty(result?.access_token))
            {
                _accessTokenCache.Set(result.access_token);
                return result.access_token;
            }

            if (result?.error == "authorization_pending")
            {
                await Task.Delay(5000, ct);
                continue;
            }

            throw new InvalidOperationException($"Auth failed: {result?.error}");
        }

        throw new TimeoutException("Device code expired");
    }

    public async Task<CopilotTokenResponse> GetCopilotTokenAsync(CancellationToken ct = default)
    {
        var cached = _tokenCache.GetValidToken();
        if (cached != null) return cached;

        var accessToken = _accessTokenCache.Get()
            ?? throw new UnauthorizedAccessException("Not authenticated");

        using var request = new HttpRequestMessage(HttpMethod.Get, CopilotTokenUrl);
        request.Headers.Add("Authorization", $"Bearer {accessToken}");
        request.Headers.Add("Accept", "application/json");
        request.Headers.Add("Editor-Version", "vscode/1.107.0");
        request.Headers.Add("User-Agent", "GitHubCopilotChat/0.35.0");

        var response = await HttpClient.SendAsync(request, ct);
        response.EnsureSuccessStatusCode();

        var result = await response.Content.ReadFromJsonAsync<CopilotTokenDto>(ct);
        var baseUrl = ExtractBaseUrl(result!.token);
        var token = new CopilotTokenResponse
        {
            Token = result.token,
            ExpiresAt = result.expires_at,
            BaseUrl = baseUrl
        };

        _tokenCache.SaveToken(token);
        return token;
    }

    public bool IsAuthenticated() =>
        _accessTokenCache.Get() != null && _tokenCache.GetValidToken() != null;

    public void Logout()
    {
        _accessTokenCache.Clear();
        _tokenCache.Clear();
    }

    private static string ExtractBaseUrl(string token)
    {
        var match = System.Text.RegularExpressions.Regex.Match(
            token,
            @"proxy-ep=([^;\s]+)",
            System.Text.RegularExpressions.RegexOptions.IgnoreCase
        );

        if (match.Success)
        {
            var proxyEp = match.Groups[1].Value
                .Replace("https://", "")
                .Replace("proxy.", "api.");
            return $"https://{proxyEp}";
        }

        return "https://api.individual.githubcopilot.com";
    }

    // Internal DTOs
    private sealed record DeviceCodeDto(
        string device_code,
        string user_code,
        string verification_uri,
        int expires_in,
        int interval
    );

    private sealed record AccessTokenDto(
        string? access_token,
        string? error
    );

    private sealed record CopilotTokenDto(
        string token,
        long expires_at
    );
}