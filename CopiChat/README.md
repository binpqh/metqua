# 🤖 CopiChat - GitHub Copilot Chat API
Clean GitHub Copilot Chat API with **tool calling & streaming** support.
## ⚡ Quick Start
```bash
dotnet run
# Visit: http://localhost:5000/scalar
```
## 🎯 Features
- ✅ GitHub device flow auth
- ✅ Auto token refresh
- ✅ Chat completions
- ✅ **Streaming (SSE)**
- ✅ **Tool calling**
- ✅ Model listing
## 📋 API
```
Auth:    POST /api/auth/device-code, /api/auth/poll
Chat:    POST /api/chat/completions, /api/chat/completions/stream ✨
Tools:   GET  /api/tools
Models:  GET  /api/models
```
## 🛠️ Tools
- read_file - Read file contents
- list_files - List directory
- search_files - Search by pattern
## 📊 Stats
Files: 16 | Lines: ~1,275 | Build: ✅
See TOOL_CALLING.md for complete docs!
