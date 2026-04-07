using CopiChat.Core.Caching;
using CopiChat.Core.Models;
using CopiChat.Features.Auth;
using CopiChat.Features.Chat;
using CopiChat.Features.Models;
using CopiChat.Features.Tools;
using Scalar.AspNetCore;

var builder = WebApplication.CreateBuilder(args);

// OpenAPI
builder.Services.AddOpenApi();

// Caching - Singleton state
builder.Services.AddSingleton<ICache<string>>(new MemoryCache<string>());
builder.Services.AddSingleton<ICache<CopilotTokenResponse>>(new MemoryCache<CopilotTokenResponse>());
builder.Services.AddSingleton<ITokenCache, TokenCache>();

// Tool execution - Singleton (workspace root doesn't change)
var workspaceRoot = builder.Environment.ContentRootPath;
builder.Services.AddSingleton<IToolExecutor>(new FileToolExecutor(workspaceRoot));
builder.Services.AddSingleton<IToolService, ToolService>();

// Feature services - Scoped
builder.Services.AddScoped<IAuthService, AuthService>();
builder.Services.AddScoped<IChatService, ChatService>();
builder.Services.AddScoped<IStreamingChatService, StreamingChatService>();
builder.Services.AddScoped<IModelService, ModelService>();

var app = builder.Build();

// Development tools
if (app.Environment.IsDevelopment())
{
    app.MapOpenApi();
    app.MapScalarApiReference();
}

app.UseHttpsRedirection();

// Map feature endpoints - VSA pattern
app.MapGroup("/api/auth").MapAuthEndpoints();
app.MapGroup("/api/chat").MapChatEndpoints();
app.MapGroup("/api/models").MapModelEndpoints();
app.MapGroup("/api/tools").MapToolEndpoints();

app.Run();

