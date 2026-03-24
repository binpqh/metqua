# Windows Server Core container — requires Windows CI runner with process isolation.
# Build: docker build -f tests/sandbox/docker/windows-ltsc/Dockerfile .
FROM mcr.microsoft.com/windows/servercore:ltsc2022

SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop';"]

# Copy the pre-built Windows binary.
COPY dist/simple-cli.exe C:\Windows\System32\simple-cli.exe

CMD ["simple-cli", "--version"]
