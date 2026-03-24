// simple-cli — long-life session CLI application.
// Constitution Principle VIII: CGO_ENABLED=0 static build.
package main

import "github.com/your-org/simple-cli/cmd"

func main() {
	cmd.Execute()
}
