import asyncio
from langchain_ollama import ChatOllama

async def main():
    # Create an instance of the ChatOllama class
    ollama = ChatOllama(model="qwen3.5:0.8b")

    # Define a prompt to send to the model
    prompt = "What is the capital of France?"

    # Get a response from the model
    response = await ollama.ainvoke(prompt)

    # Print the response
    print("Response from Ollama:", response)

if __name__ == "__main__":
    asyncio.run(main())