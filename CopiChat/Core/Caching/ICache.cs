namespace CopiChat.Core.Caching;

/// <summary>
/// Generic thread-safe cache interface
/// Follows Interface Segregation Principle
/// </summary>
/// <typeparam name="T">Cached value type</typeparam>
public interface ICache<T> where T : class
{
    T? Get();
    void Set(T value);
    void Clear();
}

/// <summary>
/// Thread-safe in-memory cache implementation
/// </summary>
/// <typeparam name="T">Cached value type</typeparam>
public sealed class MemoryCache<T> : ICache<T> where T : class
{
    private T? _value;
    private readonly Lock _lock = new();

    public T? Get()
    {
        lock (_lock)
        {
            return _value;
        }
    }

    public void Set(T value)
    {
        lock (_lock)
        {
            _value = value;
        }
    }

    public void Clear()
    {
        lock (_lock)
        {
            _value = null;
        }
    }
}