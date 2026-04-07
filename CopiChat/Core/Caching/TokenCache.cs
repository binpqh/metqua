using CopiChat.Core.Models;

namespace CopiChat.Core.Caching;

/// <summary>
/// Token cache with expiration checking
/// Single Responsibility: Manage Copilot token lifecycle
/// </summary>
public interface ITokenCache
{
    CopilotTokenResponse? GetValidToken();
    void SaveToken(CopilotTokenResponse token);
    void Clear();
}

public sealed class TokenCache : ITokenCache
{
    private readonly ICache<CopilotTokenResponse> _cache;
    private const int ExpiryBufferSeconds = 60;

    public TokenCache(ICache<CopilotTokenResponse> cache)
    {
        _cache = cache;
    }

    public CopilotTokenResponse? GetValidToken()
    {
        var token = _cache.Get();
        if (token == null || IsExpired(token))
        {
            _cache.Clear();
            return null;
        }

        return token;
    }

    public void SaveToken(CopilotTokenResponse token)
    {
        _cache.Set(token);
    }

    public void Clear()
    {
        _cache.Clear();
    }

    private static bool IsExpired(CopilotTokenResponse token)
    {
        var expiryTime = DateTimeOffset.FromUnixTimeSeconds(token.ExpiresAt);
        var bufferTime = expiryTime.AddSeconds(-ExpiryBufferSeconds);
        return DateTimeOffset.UtcNow >= bufferTime;
    }
}