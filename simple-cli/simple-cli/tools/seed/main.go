package main

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "time"

    "github.com/binpqh/simple-cli/internal/provider"
    "github.com/binpqh/simple-cli/internal/tokenstore"
)

func configDir() string {
    if runtime.GOOS == "windows" {
        if d := os.Getenv("APPDATA"); d != "" {
            return filepath.Join(d, "simple-cli")
        }
    }
    if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
        return filepath.Join(d, "simple-cli")
    }
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".config", "simple-cli")
}

func main() {
    dir := configDir()
    path := tokenstore.PathForConfigDir(dir)
    fmt.Println("Token path:", path)

    store := tokenstore.NewFileTokenStore(path)
    tok := &provider.TokenSet{
        Provider:     "mock",
        AccessToken:  "mock-access-token",
        RefreshToken: "mock-refresh-token",
        Expiry:       time.Now().Add(8766 * time.Hour), // ~1 year
        UserID:       "mock-user",
        TokenType:    "Bearer",
    }
    if err := store.Set(context.Background(), "mock", tok); err != nil {
        fmt.Fprintln(os.Stderr, "Error:", err)
        os.Exit(1)
    }
    fmt.Println("Token seeded successfully.")
    // read back to verify
    out, _ := os.ReadFile(path)
    fmt.Println("File contents:", string(out))
}
