
from mcp.server.fastmcp import FastMCP
from typing import Literal, Annotated
from pydantic import BaseModel


mcp = FastMCP(
    name="Hello MCP",
    instructions=(
        "My first MCP server - Hello World with tools, "
        "resources and prompts"
    )
)



# tools

@mcp.tool()
def hello_world(name: str = "World") -> str:
    """Greet someone in a friendly way."""
    return f"Hello {name}, I'm an MCP server running in Python"


@mcp.tool()
def add(a: int, b: int) -> int:
    """Add 2 number together."""
    return a + b


if __name__ == "__main__":
    mcp.run()
