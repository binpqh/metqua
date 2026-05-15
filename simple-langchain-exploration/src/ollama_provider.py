from typing import Optional
import os
from langchain_ollama import ChatOllama
from langchain_core.messages import AIMessage
from .llm_provider import LLMProvider

class OllamaProvider(LLMProvider):
    def __init__(self, llm : ChatOllama, model_name: str, base_url: Optional[str] = None, timeout: int = 30):
        """
        Simple Ollama provider.
        - model_name: name of the Ollama model to use
        - base_url: optional base URL for the Ollama server (defaults to env OLLAMA_BASE_URL or http://localhost:11434)
        - timeout: request timeout in seconds
        """
        self.model_name = model_name
        self.base_url = (base_url or os.getenv("OLLAMA_BASE_URL", "http://localhost:11434")).rstrip("/")
        self.timeout = timeout
        self.llm = llm

    def generate_response(self, prompt: str) -> AIMessage:
        """
        Calls an Ollama-compatible HTTP endpoint and returns the generated text.
        """
        return self.llm.invoke(prompt)