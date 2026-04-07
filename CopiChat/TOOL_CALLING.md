# Tool Calling Feature - Documentation

## 🎯 Overview

GitHub Copilot tool calling cho phép AI:
- 📁 Read files from workspace
- 📂 List directory contents
- 🔍 Search for files
- 🔄 Execute tools và tiếp tục conversation

## 🚀 Quick Start

### 1. Get Available Tools

```bash
GET /api/tools
```

**Response:**
```json
{
  "isSuccess": true,
  "data": [
    {
      "type": "function",
      "name": "read_file",
      "description": "Read contents of a file from the local workspace",
      "parameters": {
        "type": "object",
        "properties": {
          "path": {"type": "string", "description": "Relative path to the file"},
          "startLine": {"type": "number", "description": "Start line (optional)"},
          "endLine": {"type": "number", "description": "End line (optional)"}
        },
        "required": ["path"]
      }
    },
    {
      "name": "list_files",
      "description": "List files in a directory"
    },
    {
      "name": "search_files",
      "description": "Search for files matching a pattern"
    }
  ]
}
```

### 2. Stream Chat with Tools

```bash
POST /api/chat/completions/stream
Content-Type: application/json

{
  "model": "gpt-4o-2024-08-06",
  "messages": [
    {
      "role": "user",
      "content": "Read the file Program.cs and show me the Main method"
    }
  ],
  "tools": [
    {
      "type": "function",
      "name": "read_file",
      "description": "Read file contents",
      "parameters": {
        "type": "object",
        "properties": {
          "path": {"type": "string"}
        },
        "required": ["path"]
      }
    }
  ],
  "stream": true
}
```

## 📡 Server-Sent Events (SSE)

### Event Types

#### 1. tool_call_start
AI bắt đầu gọi tool:
```
event: tool_call_start
data: {"type":"tool_call_start","toolCall":{"id":"fc_123","name":"read_file"}}
```

#### 2. tool_call_delta
Streaming tool arguments:
```
event: tool_call_delta
data: {"type":"tool_call_delta","delta":"{\"path\":"}

event: tool_call_delta
data: {"type":"tool_call_delta","delta":"\"Program.cs\"}"}
```

#### 3. tool_call_done
Tool execution completed:
```
event: tool_call_done
data: {"type":"tool_call_done","toolCall":{"name":"read_file","arguments":"{\"path\":\"Program.cs\"}"}}
```

#### 4. text_delta
AI response text streaming:
```
event: text_delta
data: {"type":"text_delta","delta":"The Main "}

event: text_delta
data: {"type":"text_delta","delta":"method is "}
```

#### 5. done
Conversation complete:
```
event: done
data: {"type":"done"}
```

## 🔧 Available Tools

### read_file
Read contents of a file from workspace.

**Parameters:**
- `path` (string, required) - Relative path to file
- `startLine` (number, optional) - Start line (1-indexed)
- `endLine` (number, optional) - End line (1-indexed)

**Example:**
```json
{
  "name": "read_file",
  "arguments": {
    "path": "Program.cs",
    "startLine": 1,
    "endLine": 50
  }
}
```

**Security:**
- ✅ Path traversal prevention
- ✅ 10 MB file size limit
- ✅ Workspace root restriction

### list_files
List files and directories in a directory.

**Parameters:**
- `directory` (string, optional) - Directory path (default: workspace root)

**Example:**
```json
{
  "name": "list_files",
  "arguments": {
    "directory": "Features"
  }
}
```

### search_files
Search for files matching a pattern.

**Parameters:**
- `pattern` (string, required) - File pattern (e.g., `*.cs`, `Program.*`)
- `directory` (string, optional) - Search directory (default: workspace root)

**Example:**
```json
{
  "name": "search_files",
  "arguments": {
    "pattern": "*.cs",
    "directory": "Features"
  }
}
```

## 💻 Client-Side Usage

### JavaScript/TypeScript

```javascript
const eventSource = new EventSource('http://localhost:5000/api/chat/completions/stream', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    model: 'gpt-4o-2024-08-06',
    messages: [
      { role: 'user', content: 'Read Program.cs and explain it' }
    ],
    tools: [
      {
        type: 'function',
        name: 'read_file',
        description: 'Read file contents',
        parameters: {
          type: 'object',
          properties: {
            path: { type: 'string' }
          },
          required: ['path']
        }
      }
    ],
    stream: true
  })
});

eventSource.addEventListener('text_delta', (e) => {
  const data = JSON.parse(e.data);
  console.log(data.delta); // Stream text output
});

eventSource.addEventListener('tool_call_start', (e) => {
  const data = JSON.parse(e.data);
  console.log(`[Tool] ${data.toolCall.name} started`);
});

eventSource.addEventListener('tool_call_done', (e) => {
  const data = JSON.parse(e.data);
  console.log(`[Tool] ${data.toolCall.name} completed`);
});

eventSource.addEventListener('done', () => {
  console.log('Stream complete');
  eventSource.close();
});
```

### C# Client

```csharp
using var httpClient = new HttpClient();
var request = new HttpRequestMessage(HttpMethod.Post, "http://localhost:5000/api/chat/completions/stream")
{
    Content = JsonContent.Create(new
    {
        model = "gpt-4o-2024-08-06",
        messages = new[]
        {
            new { role = "user", content = "List all C# files in Features folder" }
        },
        tools = new[]
        {
            new
            {
                type = "function",
                name = "list_files",
                description = "List files",
                parameters = new
                {
                    type = "object",
                    properties = new
                    {
                        directory = new { type = "string" }
                    }
                }
            }
        },
        stream = true
    })
};

var response = await httpClient.SendAsync(request, HttpCompletionOption.ResponseHeadersRead);
var stream = await response.Content.ReadAsStreamAsync();
using var reader = new StreamReader(stream);

while (!reader.EndOfStream)
{
    var line = await reader.ReadLineAsync();
    if (line?.StartsWith("event: ") == true)
    {
        var eventType = line[7..];
        var dataLine = await reader.ReadLineAsync();
        
        if (dataLine?.StartsWith("data: ") == true)
        {
            var data = dataLine[6..];
            Console.WriteLine($"[{eventType}] {data}");
        }
    }
}
```

## 🔒 Security Features

### Path Traversal Prevention
```csharp
// ✅ Safe
read_file({ path: "Program.cs" })
read_file({ path: "Features/Chat/ChatService.cs" })

// ❌ Blocked
read_file({ path: "../../../etc/passwd" })
read_file({ path: "C:\\Windows\\System32\\config" })
```

### File Size Limits
- Max file size: **10 MB**
- Prevents memory exhaustion
- Returns error for large files

### Workspace Restriction
- Tools can only access files **within workspace root**
- Full path normalization and validation
- `UnauthorizedAccessException` for violations

## 🎯 Example Conversation Flow

### User Request
```
"Read Program.cs and explain the DI setup"
```

### AI Response (with tool calling)
```
1. event: tool_call_start
   → AI decides to read Program.cs

2. event: tool_call_delta (streaming args)
   → {"path": "Program.cs"}

3. event: tool_call_done
   → Tool executed, file content retrieved

4. event: text_delta (streaming response)
   → "The Program.cs file sets up dependency injection..."
   → "It registers singleton caches..."
   → "And scoped services for each feature..."

5. event: done
   → Conversation complete
```

## 📊 Architecture

```
User Request
    ↓
POST /api/chat/completions/stream
    ↓
StreamingChatService
    ↓
    ├─ Send to GitHub Copilot API
    ├─ Stream SSE events
    │   ├─ Text deltas → Client
    │   └─ Tool calls → ToolExecutor
    │       ↓
    │       Execute locally (read_file, list_files, etc.)
    │       ↓
    │       Add tool result to messages
    │       ↓
    │       Continue conversation (recursive)
    └─ Stream final response → Client
```

## ✅ Features

- ✅ **Streaming responses** - Real-time text output
- ✅ **Tool calling** - AI can call local tools
- ✅ **Auto-execution** - Tools execute automatically
- ✅ **Recursive conversation** - AI continues after tool result
- ✅ **Security** - Path traversal prevention, size limits
- ✅ **Type-safe** - Strongly-typed events
- ✅ **Clean API** - Simple SSE format

## 🧪 Test Example

```bash
# Using curl with SSE
curl -N -X POST http://localhost:5000/api/chat/completions/stream \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-2024-08-06",
    "messages": [
      {"role": "user", "content": "Show me all C# files in the project"}
    ],
    "tools": [
      {
        "type": "function",
        "name": "search_files",
        "description": "Search files",
        "parameters": {
          "type": "object",
          "properties": {
            "pattern": {"type": "string"}
          },
          "required": ["pattern"]
        }
      }
    ],
    "stream": true
  }'
```

**Expected Output:**
```
event: tool_call_start
data: {"type":"tool_call_start","toolCall":{"id":"...","name":"search_files"}}

event: tool_call_delta
data: {"type":"tool_call_delta","delta":"{\"pattern\":\"*.cs\"}"}

event: tool_call_done
data: {"type":"tool_call_done","toolCall":{"name":"search_files"}}

event: text_delta
data: {"type":"text_delta","delta":"Here are all the C# files:\n"}

event: text_delta
data: {"type":"text_delta","delta":"- Program.cs\n- Core/Models/Contracts.cs\n..."}

event: done
data: {"type":"done"}
```

## 📋 API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/tools` | GET | List available tools |
| `/api/chat/completions` | POST | Regular chat (no tools) |
| `/api/chat/completions/stream` | POST | Streaming chat with tools ✨ |

## 🎯 Summary

**New Feature:** ✅ Tool Calling + Streaming  
**Security:** ✅ Path validation, size limits  
**Performance:** ✅ Streaming for real-time UX  
**Architecture:** ✅ Clean VSA implementation  

**Status:** Ready to use! 🚀

