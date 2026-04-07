namespace CopiChat.Core.Models;

/// <summary>
/// Generic API response wrapper following .NET conventions
/// </summary>
/// <typeparam name="T">Response data type</typeparam>
public sealed class ApiResponse<T>
{
    public bool IsSuccess { get; init; }
    public string Message { get; init; } = string.Empty;
    public T? Data { get; init; }

    public static ApiResponse<T> Success(T data, string message = "") =>
        new() { IsSuccess = true, Data = data, Message = message };

    public static ApiResponse<T> Failure(string message) =>
        new() { IsSuccess = false, Message = message };
}