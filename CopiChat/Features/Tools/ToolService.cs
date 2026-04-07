using System.Text.Json;
using CopiChat.Core.Models;

namespace CopiChat.Features.Tools;

/// <summary>
/// Tool registry service
/// Provides predefined tools for GitHub Copilot
/// </summary>
public interface IToolService
{
    IReadOnlyList<Tool> GetAvailableTools();
    Tool? GetTool(string name);
}

public sealed class ToolService : IToolService
{
    private static readonly List<Tool> PredefinedTools = new()
    {
        new Tool
        {
            Type = "function",
            Name = "read_file",
            Description = "Read contents of a file from the local workspace",
            Parameters = JsonDocument.Parse("""
                {
                    "type": "object",
                    "properties": {
                        "path": {
                            "type": "string",
                            "description": "Relative path to the file to read"
                        },
                        "startLine": {
                            "type": "number",
                            "description": "Starting line number (1-indexed, optional)"
                        },
                        "endLine": {
                            "type": "number",
                            "description": "Ending line number (1-indexed, optional)"
                        }
                    },
                    "required": ["path"]
                }
                """).RootElement,
            Strict = false
        },
        new Tool
        {
            Type = "function",
            Name = "list_files",
            Description = "List files and directories in a directory",
            Parameters = JsonDocument.Parse("""
                {
                    "type": "object",
                    "properties": {
                        "directory": {
                            "type": "string",
                            "description": "Path to directory to list (default: workspace root)"
                        }
                    },
                    "required": []
                }
                """).RootElement,
            Strict = false
        },
        new Tool
        {
            Type = "function",
            Name = "search_files",
            Description = "Search for files matching a pattern in the workspace",
            Parameters = JsonDocument.Parse("""
                {
                    "type": "object",
                    "properties": {
                        "pattern": {
                            "type": "string",
                            "description": "File name pattern (e.g., *.cs, Program.*)"
                        },
                        "directory": {
                            "type": "string",
                            "description": "Directory to search (default: workspace root)"
                        }
                    },
                    "required": ["pattern"]
                }
                """).RootElement,
            Strict = false
        }
    };

    public IReadOnlyList<Tool> GetAvailableTools() => PredefinedTools;

    public Tool? GetTool(string name) =>
        PredefinedTools.FirstOrDefault(t => t.Name == name);
}

