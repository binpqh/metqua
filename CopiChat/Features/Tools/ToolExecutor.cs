using System.Text.Json;
using CopiChat.Core.Models;

namespace CopiChat.Features.Tools;

/// <summary>
/// Tool executor interface
/// </summary>
public interface IToolExecutor
{
    Task<ToolResult> ExecuteAsync(string toolName, JsonElement arguments, CancellationToken ct = default);
    bool CanExecute(string toolName);
}

/// <summary>
/// File system tool executor with security checks
/// </summary>
public sealed class FileToolExecutor : IToolExecutor
{
    private readonly string _workspaceRoot;
    private const long MaxFileSizeBytes = 10 * 1024 * 1024; // 10 MB

    public FileToolExecutor(string workspaceRoot)
    {
        _workspaceRoot = Path.GetFullPath(workspaceRoot);
    }

    public bool CanExecute(string toolName) =>
        toolName is "read_file" or "list_files" or "search_files";

    public async Task<ToolResult> ExecuteAsync(string toolName, JsonElement arguments, CancellationToken ct = default)
    {
        return toolName switch
        {
            "read_file" => await ReadFileAsync(arguments, ct),
            "list_files" => await ListFilesAsync(arguments, ct),
            "search_files" => await SearchFilesAsync(arguments, ct),
            _ => new ToolResult
            {
                CallId = string.Empty,
                Output = $"Unknown tool: {toolName}",
                IsError = true
            }
        };
    }

    private async Task<ToolResult> ReadFileAsync(JsonElement args, CancellationToken ct)
    {
        try
        {
            var path = args.GetProperty("path").GetString() 
                ?? throw new ArgumentException("path is required");

            var fullPath = GetSafePath(path);
            
            // Security: Check file size
            var fileInfo = new FileInfo(fullPath);
            if (fileInfo.Length > MaxFileSizeBytes)
            {
                return ErrorResult($"File too large: {fileInfo.Length / 1024 / 1024} MB (max: 10 MB)");
            }

            var content = await File.ReadAllTextAsync(fullPath, ct);

            // Handle line ranges
            if (args.TryGetProperty("startLine", out var startProp) && 
                args.TryGetProperty("endLine", out var endProp))
            {
                var lines = content.Split('\n');
                var startLine = Math.Max(0, startProp.GetInt32() - 1);
                var endLine = Math.Min(lines.Length - 1, endProp.GetInt32() - 1);
                
                content = string.Join('\n', lines.Skip(startLine).Take(endLine - startLine + 1));
            }

            return SuccessResult(content);
        }
        catch (Exception ex)
        {
            return ErrorResult($"Error reading file: {ex.Message}");
        }
    }

    private async Task<ToolResult> ListFilesAsync(JsonElement args, CancellationToken ct)
    {
        try
        {
            var directory = args.GetProperty("directory").GetString() ?? ".";
            var fullPath = GetSafePath(directory);

            if (!Directory.Exists(fullPath))
            {
                return ErrorResult($"Directory not found: {directory}");
            }

            var files = Directory.GetFileSystemEntries(fullPath)
                .Select(f => Path.GetRelativePath(_workspaceRoot, f))
                .OrderBy(f => f);

            var output = string.Join("\n", files);
            return SuccessResult(output);
        }
        catch (Exception ex)
        {
            return ErrorResult($"Error listing files: {ex.Message}");
        }
    }

    private async Task<ToolResult> SearchFilesAsync(JsonElement args, CancellationToken ct)
    {
        try
        {
            var pattern = args.GetProperty("pattern").GetString() ?? "*";
            var directory = args.TryGetProperty("directory", out var dirProp) 
                ? dirProp.GetString() ?? "." 
                : ".";

            var fullPath = GetSafePath(directory);
            var files = Directory.GetFiles(fullPath, pattern, SearchOption.AllDirectories)
                .Select(f => Path.GetRelativePath(_workspaceRoot, f))
                .Take(100) // Limit results
                .OrderBy(f => f);

            var output = string.Join("\n", files);
            return SuccessResult(output);
        }
        catch (Exception ex)
        {
            return ErrorResult($"Error searching files: {ex.Message}");
        }
    }

    private string GetSafePath(string relativePath)
    {
        var fullPath = Path.GetFullPath(Path.Combine(_workspaceRoot, relativePath));
        
        // Security: Prevent path traversal
        if (!fullPath.StartsWith(_workspaceRoot + Path.DirectorySeparatorChar) && 
            fullPath != _workspaceRoot)
        {
            throw new UnauthorizedAccessException($"Access denied: {relativePath}");
        }

        return fullPath;
    }

    private static ToolResult SuccessResult(string output) =>
        new() { CallId = string.Empty, Output = output, IsError = false };

    private static ToolResult ErrorResult(string message) =>
        new() { CallId = string.Empty, Output = message, IsError = true };
}

