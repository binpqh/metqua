# Simple RAG - Contact Data Quality Assistant

A Retrieval Augmented Generation (RAG) system built in C# to help identify and eliminate "dirty records" in contact management applications using local LLM (LMStudio).

## 🎯 Purpose

This RAG system helps staff officers analyze contact records imported from external systems, identifying data quality issues and providing recommendations based on documented quality standards.

## 🏗️ Architecture

```
┌─────────────────┐
│   Documents     │ (.txt, .md files)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Chunking       │ (512 words, 64 overlap)
└────────┬────────┘
         │
         ▼
┌─────────────────┐      ┌──────────────┐
│  LMStudio       │◄─────┤  Embeddings  │
│  (Local LLM)    │      │  (768-dim)   │
└────────┬────────┘      └──────────────┘
         │
         ▼
┌─────────────────┐
│  Vector Store   │ (In-memory, cosine similarity)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Query + RAG    │ → Answer with sources
└─────────────────┘
```

## 📋 Prerequisites

1. **.NET 6 SDK or later**

   ```bash
   dotnet --version  # Should be 6.0 or higher
   ```

2. **LMStudio** running locally
   - Download from: https://lmstudio.ai/
   - Load an embedding model (e.g., `nomic-embed-text`)
   - Load a completion model (e.g., `llama-2-7b`)
   - Start the local server (default: `http://localhost:1234`)

## 🚀 Quick Start

### 1. Build the project

```bash
cd simple-rag
dotnet build
```

### 2. Configure LMStudio

Edit `appsettings.json` if your LMStudio uses a different URL or models:

```json
{
  "LMStudio": {
    "BaseUrl": "http://localhost:1234",
    "EmbeddingModel": "nomic-embed-text",
    "CompletionModel": "llama-7b"
  }
}
```

### 3. Run the application

```bash
dotnet run
```

## 📖 Usage

### Interactive Mode

Run without arguments for interactive mode:

```bash
dotnet run
```

Available commands:

```
ingest <path>     - Ingest documents from a directory
query <question>  - Ask a question using RAG
stats             - Show system statistics
clear             - Clear all ingested data
help              - Show help message
exit/quit         - Exit the application
```

### Example Session

```bash
> ingest ./documents
Found 2 documents to ingest...

Processing: contact-data-quality.md
  - Created 45 chunks
  - Embedding chunks 1-10...
  - Embedding chunks 11-20...
  ...
  ✓ Completed contact-data-quality.md

INGESTION SUMMARY
Files processed: 2
Chunks created: 89
Duration: 12.34 seconds

> query What makes an email address invalid?

Embedding question...
Searching for top 5 relevant chunks...
Generating answer...

═══════════════════════════════════════════════════════════
QUESTION:
What makes an email address invalid?

ANSWER:
An invalid email address typically has one or more of these issues:
- Missing or multiple @ symbols
- No local part before the @
- No domain after the @
...

RETRIEVED CHUNKS (5):
[1] contact-data-quality.md (Score: 0.912)
    A valid email address must follow these criteria: Contains exactly...

═══════════════════════════════════════════════════════════
```

### Command-Line Mode

Execute single commands:

```bash
# Ingest documents
dotnet run ingest ./documents

# Query
dotnet run query "What phone number patterns indicate dirty data?"

# Show statistics
dotnet run stats
```

## 🗂️ Project Structure

```
simple-rag/
├── Core/
│   ├── Document.cs          # Document model
│   ├── Chunk.cs             # Chunk model with embeddings
│   └── RagConfig.cs         # Configuration models
├── Services/
│   ├── LMStudioClient.cs    # LMStudio API client
│   ├── DocumentChunker.cs   # Text chunking logic
│   ├── RetrievalPipeline.cs # RAG orchestration
│   ├── DocumentIngestionService.cs
│   └── ContactDataQualityService.cs
├── Storage/
│   └── VectorStore.cs       # In-memory vector storage
├── documents/               # Knowledge base documents
│   ├── contact-data-quality.md
│   └── import-quality-guidelines.md
├── Program.cs               # CLI interface
├── appsettings.json         # Configuration
└── SimpleRag.csproj         # Project file
```

## 🎓 How It Works

### 1. Document Ingestion

Documents are split into overlapping chunks (512 words with 64-word overlap) and converted to 768-dimensional vectors using LMStudio's embedding model.

```csharp
// Documents → Chunks → Embeddings → Vector Store
var chunks = chunker.ChunkDocument(document);
var embeddings = await lmStudio.GetEmbeddingsBatchAsync(texts);
vectorStore.AddChunks(chunks);
```

### 2. Query Processing

Questions are embedded and compared against stored chunks using cosine similarity. The top 5 most relevant chunks provide context for the LLM to generate an answer.

```csharp
// Query → Embedding → Search → Context → LLM → Answer
var embedding = await lmStudio.GetEmbeddingAsync(question);
var relevant = vectorStore.SearchTopK(embedding, 5);
var answer = await lmStudio.GetCompletionAsync(question, context);
```

### 3. Contact Analysis

The `ContactDataQualityService` can analyze contact records:

```csharp
var service = new ContactDataQualityService(pipeline, lmStudio, vectorStore);

var contact = new ContactRecord
{
    Id = "12345",
    Name = "John Smith",
    Email = "john@@company.com",  // Invalid!
    Phone = "111-1111",           // Suspicious!
    Company = "Test Company"      // Generic!
};

var result = await service.AnalyzeContactRecordAsync(contact);
result.PrintToConsole();
// Output: Risk Score: 85/100 (HIGH RISK)
// Issues: Multiple @ symbols, Sequential phone number, Test company name
```

## ⚙️ Configuration

### RAG Settings

Adjust in `appsettings.json`:

```json
{
  "RAG": {
    "MaxChunksPerQuery": 5, // Number of chunks to retrieve
    "ChunkSize": 512, // Words per chunk
    "ChunkOverlap": 64, // Overlapping words
    "MaxDocuments": 100, // Maximum documents
    "DocumentsPath": "./documents"
  }
}
```

### Performance Tuning

- **Chunk Size**: Larger chunks (1024) for more context, smaller (256) for precision
- **Chunk Overlap**: More overlap (128) for better continuity, less for speed
- **MaxChunksPerQuery**: More chunks (10) for comprehensive answers, fewer (3) for speed
- **Batch Size**: Modify in `DocumentIngestionService.cs` for embedding batches

## 🔧 Integration with Contact Management App

### Option 1: Reference as Library

```csharp
// In your contact management app
using SimpleRag.Services;

var config = LoadRagConfig();
var lmStudio = new LMStudioClient(config.LMStudio);
var vectorStore = new VectorStore(768);
var pipeline = new RetrievalPipeline(vectorStore, lmStudio, 5);

// Analyze imported contacts
var qualityService = new ContactDataQualityService(pipeline, lmStudio, vectorStore);
var result = await qualityService.AnalyzeContactRecordAsync(importedContact);

if (result.RiskScore > 70)
{
    // Flag for manual review
    await FlagForReview(importedContact, result.Issues);
}
```

### Option 2: Add Web API

Create an ASP.NET Core Web API wrapper (future enhancement) for HTTP-based integration.

## 📊 Sample Documents Included

1. **contact-data-quality.md**: Standards for email, phone, name, and company validation
2. **import-quality-guidelines.md**: Guidelines for handling dirty data from external systems

Add your own documents to the `documents/` folder and re-run `ingest`.

## 🐛 Troubleshooting

### "Cannot connect to LMStudio"

- Ensure LMStudio is running
- Check the URL in `appsettings.json` (default: `http://localhost:1234`)
- Verify embedding and completion models are loaded in LMStudio

### "No embedding data returned"

- Check that an embedding model is loaded in LMStudio
- Try a different embedding model
- Verify the model name in `appsettings.json` matches LMStudio

### Memory Issues

- Reduce `ChunkSize` and `MaxChunksPerQuery` in config
- Process fewer documents at once
- Reduce embedding batch size in `DocumentIngestionService.cs`

## 🚧 Future Enhancements

- [ ] FAISS integration for better performance
- [ ] Persistent storage (save/load vector index)
- [ ] Web API for remote access
- [ ] PDF/Word document support
- [ ] Batch contact analysis CLI command
- [ ] Quality score trend analysis
- [ ] Custom quality rules configuration

## 📝 License

This is a sample implementation for educational purposes. Adapt as needed for your use case.

## 🤝 Contributing

This is a foundational implementation. To extend:

1. Add new document types in `DocumentIngestionService`
2. Enhance quality analysis in `ContactDataQualityService`
3. Replace `VectorStore` with FAISS for production
4. Add structured output parsing for LLM responses
5. Implement persistence layer

---

**Built with specification-driven development and AI assistance.**
