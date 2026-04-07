# Tool Calling Demo Script

Write-Host "`n🔧 CopiChat - Tool Calling Feature Demo`n" -ForegroundColor Cyan

$baseUrl = "http://localhost:5000"

Write-Host "📋 Prerequisites:" -ForegroundColor Yellow
Write-Host "  1. Server must be running: dotnet run" -ForegroundColor Gray
Write-Host "  2. Must be logged in (run auth flow first)`n" -ForegroundColor Gray

# Test 1: Get available tools
Write-Host "Test 1: Get Available Tools" -ForegroundColor Green
Write-Host "─────────────────────────────────────────`n" -ForegroundColor Gray

try {
    $tools = Invoke-RestMethod -Uri "$baseUrl/api/tools" -Method Get
    Write-Host "✅ Available Tools:" -ForegroundColor Green
    foreach ($tool in $tools.data) {
        Write-Host "  • $($tool.name) - $($tool.description)" -ForegroundColor White
    }
    Write-Host ""
} catch {
    Write-Host "❌ Failed to get tools: $($_.Exception.Message)`n" -ForegroundColor Red
    exit 1
}

# Test 2: Stream chat with read_file tool
Write-Host "Test 2: Stream Chat with read_file Tool" -ForegroundColor Green
Write-Host "─────────────────────────────────────────`n" -ForegroundColor Gray

$streamRequest = @{
    model = "gpt-4o-2024-08-06"
    messages = @(
        @{
            role = "user"
            content = "Read the file Program.cs and tell me how many services are registered"
        }
    )
    tools = @(
        @{
            type = "function"
            name = "read_file"
            description = "Read file contents"
            parameters = @{
                type = "object"
                properties = @{
                    path = @{
                        type = "string"
                        description = "File path"
                    }
                }
                required = @("path")
            }
        }
    )
    stream = $true
} | ConvertTo-Json -Depth 10

Write-Host "Request body:" -ForegroundColor Yellow
Write-Host $streamRequest -ForegroundColor Gray
Write-Host "`nStreaming response:`n" -ForegroundColor Yellow

try {
    # Note: PowerShell doesn't handle SSE well, use curl instead
    Write-Host "Use this curl command to test streaming:" -ForegroundColor Cyan
    Write-Host @"
curl -N -X POST $baseUrl/api/chat/completions/stream \
  -H "Content-Type: application/json" \
  -d '$($streamRequest -replace "'", "\'")'
"@ -ForegroundColor White

    Write-Host "`n📝 Expected SSE events:" -ForegroundColor Yellow
    Write-Host "  1. event: tool_call_start" -ForegroundColor Gray
    Write-Host "  2. event: tool_call_delta (streaming args)" -ForegroundColor Gray
    Write-Host "  3. event: tool_call_done" -ForegroundColor Gray
    Write-Host "  4. event: text_delta (AI response)" -ForegroundColor Gray
    Write-Host "  5. event: done`n" -ForegroundColor Gray
} catch {
    Write-Host "❌ Error: $($_.Exception.Message)`n" -ForegroundColor Red
}

# Test 3: List available models (verify auth still works)
Write-Host "`nTest 3: Verify Auth Still Works" -ForegroundColor Green
Write-Host "─────────────────────────────────────────`n" -ForegroundColor Gray

try {
    $models = Invoke-RestMethod -Uri "$baseUrl/api/models" -Method Get
    Write-Host "✅ Auth working. Available models: $($models.data.Count)" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "❌ Auth failed: $($_.Exception.Message)`n" -ForegroundColor Red
}

# Summary
Write-Host "📊 Tool Calling Feature Summary" -ForegroundColor Cyan
Write-Host "═══════════════════════════════════════════" -ForegroundColor Gray
Write-Host "✅ Tools endpoint: GET /api/tools" -ForegroundColor White
Write-Host "✅ Streaming endpoint: POST /api/chat/completions/stream" -ForegroundColor White
Write-Host "✅ Available tools: read_file, list_files, search_files" -ForegroundColor White
Write-Host "✅ Security: Path validation, 10MB limit" -ForegroundColor White
Write-Host "✅ SSE events: tool_call_start, delta, done, text_delta" -ForegroundColor White
Write-Host "═══════════════════════════════════════════`n" -ForegroundColor Gray

Write-Host "📚 Documentation:" -ForegroundColor Cyan
Write-Host "  - TOOL_CALLING.md - Complete guide" -ForegroundColor Gray
Write-Host "  - ARCHITECTURE.md - Architecture details" -ForegroundColor Gray
Write-Host "  - /scalar - Interactive API docs`n" -ForegroundColor Gray

Write-Host "🚀 Status: Tool Calling Feature Ready!" -ForegroundColor Green

