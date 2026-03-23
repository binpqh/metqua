using SimpleRag.Storage;

namespace SimpleRag.Services;

/// <summary>
/// Service for analyzing contact data quality using RAG
/// Helps identify "dirty records" imported from external systems
/// </summary>
public class ContactDataQualityService
{
    private readonly RetrievalPipeline _pipeline;
    private readonly IModelClient _lmStudioClient;
    private readonly VectorStore _vectorStore;

    public ContactDataQualityService(
        RetrievalPipeline pipeline,
        IModelClient lmStudioClient,
        VectorStore vectorStore)
    {
        _pipeline = pipeline;
        _lmStudioClient = lmStudioClient;
        _vectorStore = vectorStore;
    }

    /// <summary>
    /// Analyze a contact record for potential quality issues
    /// </summary>
    public async Task<ContactAnalysisResult> AnalyzeContactRecordAsync(
        ContactRecord contact,
        CancellationToken cancellationToken = default)
    {
        var startTime = DateTime.UtcNow;

        // Build analysis query
        var query = BuildAnalysisQuery(contact);

        // Get relevant context from the knowledge base
        var queryEmbedding = await _lmStudioClient.GetEmbeddingAsync(query, cancellationToken);
        var searchResults = _vectorStore.SearchTopK(queryEmbedding, 5);

        if (searchResults.Count == 0)
        {
            return new ContactAnalysisResult
            {
                ContactId = contact.Id,
                RiskScore = 0,
                Issues = new List<string> { "No quality standards found in knowledge base" },
                Recommendations = new List<string> { "Please ingest data quality documentation first" },
                ProcessingTimeMs = (int)(DateTime.UtcNow - startTime).TotalMilliseconds
            };
        }

        // Build context from retrieved chunks
        var context = BuildQualityContext(searchResults);

        // Generate analysis using LLM
        var analysisPrompt = BuildAnalysisPrompt(contact, context);
        var analysis = await _lmStudioClient.GetCompletionAsync(analysisPrompt, context, cancellationToken);

        // Parse the analysis (in a real system, you'd use structured output)
        var result = ParseAnalysisResult(contact.Id, analysis);
        result.ProcessingTimeMs = (int)(DateTime.UtcNow - startTime).TotalMilliseconds;

        return result;
    }

    /// <summary>
    /// Analyze multiple contact records in batch
    /// </summary>
    public async Task<List<ContactAnalysisResult>> AnalyzeBatchAsync(
        List<ContactRecord> contacts,
        CancellationToken cancellationToken = default)
    {
        var results = new List<ContactAnalysisResult>();

        foreach (var contact in contacts)
        {
            try
            {
                var result = await AnalyzeContactRecordAsync(contact, cancellationToken);
                results.Add(result);
            }
            catch (Exception ex)
            {
                results.Add(new ContactAnalysisResult
                {
                    ContactId = contact.Id,
                    RiskScore = 100,
                    Issues = new List<string> { $"Analysis failed: {ex.Message}" },
                    Recommendations = new List<string> { "Manual review required" }
                });
            }
        }

        return results;
    }

    private string BuildAnalysisQuery(ContactRecord contact)
    {
        return $"Analyze this contact record for data quality issues: " +
               $"Name: {contact.Name}, " +
               $"Email: {contact.Email}, " +
               $"Phone: {contact.Phone}, " +
               $"Company: {contact.Company}. " +
               $"What quality problems exist?";
    }

    private string BuildQualityContext(List<SearchResult> searchResults)
    {
        var contextParts = searchResults.Select((result, index) =>
        {
            var sourceFile = Path.GetFileName(result.Chunk.SourcePath);
            return $"[Quality Standard {index + 1}] (Source: {sourceFile})\n{result.Chunk.Text}";
        });

        return string.Join("\n\n", contextParts);
    }

    private string BuildAnalysisPrompt(ContactRecord contact, string context)
    {
        return $@"Analyze this contact record for data quality issues based on the standards provided:

CONTACT RECORD:
- ID: {contact.Id}
- Name: {contact.Name}
- Email: {contact.Email}
- Phone: {contact.Phone}
- Company: {contact.Company}
- Source System: {contact.SourceSystem}

TASK:
1. Identify all data quality issues using the provided quality standards
2. Assign a risk score (0-100, where 100 is highest risk)
3. List specific issues found
4. Provide actionable recommendations

Format your response as:
RISK_SCORE: [number]
ISSUES:
- [issue 1]
- [issue 2]
RECOMMENDATIONS:
- [recommendation 1]
- [recommendation 2]";
    }

    private ContactAnalysisResult ParseAnalysisResult(string contactId, string analysis)
    {
        var result = new ContactAnalysisResult { ContactId = contactId };

        try
        {
            // Simple parsing (in production, use structured output or better parsing)
            var lines = analysis.Split('\n', StringSplitOptions.RemoveEmptyEntries);
            var section = "";

            foreach (var line in lines)
            {
                var trimmed = line.Trim();

                if (trimmed.StartsWith("RISK_SCORE:", StringComparison.OrdinalIgnoreCase))
                {
                    var scoreStr = trimmed.Substring(11).Trim().TrimEnd('%');
                    if (int.TryParse(scoreStr, out var score))
                    {
                        result.RiskScore = Math.Clamp(score, 0, 100);
                    }
                }
                else if (trimmed.StartsWith("ISSUES:", StringComparison.OrdinalIgnoreCase))
                {
                    section = "issues";
                }
                else if (trimmed.StartsWith("RECOMMENDATIONS:", StringComparison.OrdinalIgnoreCase))
                {
                    section = "recommendations";
                }
                else if (trimmed.StartsWith("-") || trimmed.StartsWith("•"))
                {
                    var item = trimmed.TrimStart('-', '•').Trim();
                    if (!string.IsNullOrEmpty(item))
                    {
                        if (section == "issues")
                            result.Issues.Add(item);
                        else if (section == "recommendations")
                            result.Recommendations.Add(item);
                    }
                }
            }

            // If parsing failed, store raw analysis
            if (result.Issues.Count == 0)
            {
                result.Issues.Add("See raw analysis");
                result.RawAnalysis = analysis;
            }
        }
        catch
        {
            result.Issues.Add("Failed to parse analysis");
            result.RawAnalysis = analysis;
        }

        return result;
    }
}

/// <summary>
/// Represents a contact record to be analyzed
/// </summary>
public class ContactRecord
{
    public string Id { get; set; } = string.Empty;
    public string Name { get; set; } = string.Empty;
    public string Email { get; set; } = string.Empty;
    public string Phone { get; set; } = string.Empty;
    public string Company { get; set; } = string.Empty;
    public string SourceSystem { get; set; } = string.Empty;
}

/// <summary>
/// Result of contact data quality analysis
/// </summary>
public class ContactAnalysisResult
{
    public string ContactId { get; set; } = string.Empty;
    public int RiskScore { get; set; }
    public List<string> Issues { get; set; } = new();
    public List<string> Recommendations { get; set; } = new();
    public string RawAnalysis { get; set; } = string.Empty;
    public int ProcessingTimeMs { get; set; }

    public void PrintToConsole()
    {
        Console.WriteLine("\n" + new string('=', 80));
        Console.WriteLine($"CONTACT ANALYSIS - ID: {ContactId}");
        Console.WriteLine(new string('=', 80));

        var riskColor = RiskScore switch
        {
            >= 70 => "HIGH RISK",
            >= 40 => "MEDIUM RISK",
            _ => "LOW RISK"
        };

        Console.WriteLine($"Risk Score: {RiskScore}/100 ({riskColor})");

        Console.WriteLine($"\nISSUES FOUND ({Issues.Count}):");
        foreach (var issue in Issues)
        {
            Console.WriteLine($"  ❌ {issue}");
        }

        Console.WriteLine($"\nRECOMMENDATIONS ({Recommendations.Count}):");
        foreach (var rec in Recommendations)
        {
            Console.WriteLine($"  ✓ {rec}");
        }

        if (!string.IsNullOrEmpty(RawAnalysis))
        {
            Console.WriteLine($"\nRAW ANALYSIS:");
            Console.WriteLine(RawAnalysis);
        }

        Console.WriteLine($"\nProcessing time: {ProcessingTimeMs}ms");
        Console.WriteLine(new string('=', 80) + "\n");
    }
}
