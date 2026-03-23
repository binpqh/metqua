using System.Text.Json;
using SimpleRag.Core;
using SimpleRag.Services;
using SimpleRag.Storage;

namespace SimpleRag;

class Program
{
    private static RagConfig? _config;
    private static IModelClient? _lmStudioClient;
    private static VectorStore? _vectorStore;
    private static DocumentChunker? _chunker;
    private static DocumentIngestionService? _ingestionService;
    private static RetrievalPipeline? _pipeline;

    static async Task<int> Main(string[] args)
    {
        Console.WriteLine("╔═══════════════════════════════════════════════════════════════╗");
        Console.WriteLine("║           Simple RAG - Contact Data Quality Assistant         ║");
        Console.WriteLine("╚═══════════════════════════════════════════════════════════════╝\n");

        // Load configuration
        if (!LoadConfiguration())
        {
            return 1;
        }

        // Initialize services
        InitializeServices();

        // Test LMStudio connection
        if (!await TestLMStudioConnection())
        {
            return 1;
        }

        // Parse command line arguments
        if (args.Length == 0)
        {
            // Interactive mode
            await RunInteractiveMode();
        }
        else
        {
            // Command mode
            await ExecuteCommand(args);
        }

        return 0;
    }

    private static bool LoadConfiguration()
    {
        try
        {
            var configPath = "appsettings.json";
            if (!File.Exists(configPath))
            {
                Console.WriteLine($"❌ Configuration file not found: {configPath}");
                return false;
            }

            var json = File.ReadAllText(configPath);
            _config = JsonSerializer.Deserialize<RagConfig>(json, new JsonSerializerOptions
            {
                PropertyNameCaseInsensitive = true
            });

            if (_config == null)
            {
                Console.WriteLine("❌ Failed to parse configuration");
                return false;
            }

            Console.WriteLine("✓ Configuration loaded");
            return true;
        }
        catch (Exception ex)
        {
            Console.WriteLine($"❌ Error loading configuration: {ex.Message}");
            return false;
        }
    }

    private static void InitializeServices()
    {
        // Choose a provider based on configuration
        if ((_config!.Provider ?? "").ToLowerInvariant() == "ollama")
        {
            _lmStudioClient = new OllamaClient(_config.Ollama);
            _vectorStore = new VectorStore(_config.Ollama.EmbeddingDimension);
        }
        else
        {
            _lmStudioClient = new LMStudioClient(_config.LMStudio);
            _vectorStore = new VectorStore(_config.LMStudio.EmbeddingDimension);
        }
        _chunker = new DocumentChunker(_config.RAG.ChunkSize, _config.RAG.ChunkOverlap);
        _ingestionService = new DocumentIngestionService(_chunker, _lmStudioClient, _vectorStore);
        _pipeline = new RetrievalPipeline(_vectorStore, _lmStudioClient, _config.RAG.MaxChunksPerQuery);

        Console.WriteLine("✓ Services initialized\n");
    }

    private static async Task<bool> TestLMStudioConnection()
    {
        var provider = (_config!.Provider ?? "lmstudio").ToLowerInvariant();
        var url = provider == "ollama" ? _config.Ollama.BaseUrl : _config.LMStudio.BaseUrl;

        Console.WriteLine($"Testing connection to {provider} at {url}...");

        var isConnected = await _lmStudioClient!.TestConnectionAsync();

        if (!isConnected)
        {
            Console.WriteLine($"❌ Cannot connect to {provider}");
            Console.WriteLine($"   Please ensure {provider} is running at {url}");
            Console.WriteLine("   and that an embedding model is loaded.");
            return false;
        }

        Console.WriteLine($"✓ {provider} connection OK\n");
        return true;
    }

    private static async Task RunInteractiveMode()
    {
        Console.WriteLine("Interactive Mode - Type 'help' for available commands\n");

        while (true)
        {
            Console.Write("> ");
            var input = Console.ReadLine()?.Trim();

            if (string.IsNullOrEmpty(input))
                continue;

            var parts = ParseCommand(input);

            if (parts[0].ToLower() == "exit" || parts[0].ToLower() == "quit")
            {
                Console.WriteLine("Goodbye!");
                break;
            }

            await ExecuteCommand(parts);
        }
    }

    private static async Task ExecuteCommand(string[] args)
    {
        try
        {
            var command = args[0].ToLower();

            switch (command)
            {
                case "help":
                    ShowHelp();
                    break;

                case "ingest":
                    if (args.Length < 2)
                    {
                        Console.WriteLine("Usage: ingest <directory_path>");
                        break;
                    }
                    await IngestCommand(args[1]);
                    break;

                case "query":
                    if (args.Length < 2)
                    {
                        Console.WriteLine("Usage: query <your question>");
                        break;
                    }
                    var question = string.Join(" ", args.Skip(1));
                    await QueryCommand(question);
                    break;

                case "stats":
                    ShowStats();
                    break;

                case "clear":
                    ClearCommand();
                    break;

                default:
                    Console.WriteLine($"Unknown command: {command}");
                    Console.WriteLine("Type 'help' for available commands");
                    break;
            }
        }
        catch (Exception ex)
        {
            Console.WriteLine($"❌ Error: {ex.Message}");
        }
    }

    private static void ShowHelp()
    {
        Console.WriteLine("\nAvailable Commands:");
        Console.WriteLine("─────────────────────────────────────────────────────────────");
        Console.WriteLine("  ingest <path>     - Ingest documents from a directory");
        Console.WriteLine("  query <question>  - Ask a question using RAG");
        Console.WriteLine("  stats             - Show system statistics");
        Console.WriteLine("  clear             - Clear all ingested data");
        Console.WriteLine("  help              - Show this help message");
        Console.WriteLine("  exit/quit         - Exit the application");
        Console.WriteLine("─────────────────────────────────────────────────────────────\n");
    }

    private static async Task IngestCommand(string path)
    {
        Console.WriteLine();
        var result = await _ingestionService!.IngestDirectoryAsync(path);
        result.PrintSummary();
    }

    private static async Task QueryCommand(string question)
    {
        var stats = _vectorStore!.GetStats();

        if (stats.TotalChunks == 0)
        {
            Console.WriteLine("❌ No documents have been ingested yet.");
            Console.WriteLine("   Use the 'ingest' command to add documents first.");
            return;
        }

        Console.WriteLine();
        var response = await _pipeline!.RetrieveAndAnswerAsync(question);
        response.PrintToConsole();
    }

    private static void ShowStats()
    {
        var stats = _vectorStore!.GetStats();

        Console.WriteLine("\n" + new string('=', 80));
        Console.WriteLine("SYSTEM STATISTICS");
        Console.WriteLine(new string('=', 80));
        Console.WriteLine($"Documents ingested:    {stats.UniqueDocuments}");
        Console.WriteLine($"Total chunks:          {stats.TotalChunks}");
        Console.WriteLine($"Vector dimension:      {stats.Dimension}");
        Console.WriteLine($"Chunks per query:      {_config!.RAG.MaxChunksPerQuery}");
        Console.WriteLine($"Chunk size (words):    {_config.RAG.ChunkSize}");
        Console.WriteLine($"Chunk overlap (words): {_config.RAG.ChunkOverlap}");
        Console.WriteLine(new string('=', 80) + "\n");
    }

    private static void ClearCommand()
    {
        _vectorStore!.Clear();
        Console.WriteLine("✓ All data cleared from vector store\n");
    }

    private static string[] ParseCommand(string input)
    {
        var parts = new List<string>();
        var current = string.Empty;
        var inQuotes = false;

        foreach (var c in input)
        {
            if (c == '"')
            {
                inQuotes = !inQuotes;
            }
            else if (c == ' ' && !inQuotes)
            {
                if (!string.IsNullOrEmpty(current))
                {
                    parts.Add(current);
                    current = string.Empty;
                }
            }
            else
            {
                current += c;
            }
        }

        if (!string.IsNullOrEmpty(current))
        {
            parts.Add(current);
        }

        return parts.ToArray();
    }
}
