
from mcp.server.fastmcp import FastMCP, Context
from mcp.server.fastmcp.resources import Resource
from mcp.server.fastmcp.prompts import Prompt
from typing import Literal, Annotated, Dict, List
from pydantic import BaseModel, Field


mcp = FastMCP(
    name="Penguin MCP",
    instructions="Provide tools, resources and prompts for integrations.",
)

# schemas


class SearchParams(BaseModel):
    query: str
    max_results: int = 5


class SearchRequest(BaseModel):
    query: str = Field(
        ...,
        description="Search Query",
    )
    filters: dict = Field(
        default_factory=dict,
        description="Additional Filters",
    )
    max_results: int = Field(10, ge=1, le=50,)
    include_sources: bool = True

# tools


@mcp.tool()
async def advanced_search(
    request: SearchRequest
) -> List[Dict]:
    """Perform advanced semantic search with filtering."""
    return [{"ok": True, }]


@mcp.tool()
async def analyze_sentiment(
    text: str,
    mode: Literal["positive", "negative", "detailed"] = "detailed"
) -> Dict:
    """ Analyze the sentiment of provided text.

    Args:
     text: The text to analyze
     mode: Analysis depth level
    """
    return dict(
        sentiment="positive",
        score=0.87,
        key_phrases=["excellent", "highly recommended",],
    )


@mcp.tool()
async def web_search(params: SearchParams, ctx: Context) -> List[Dict]:
    await ctx.info(f"Searching for: {params.query}")
    # simulate search
    return [
        {"title": f"Result for {params.query}", "url": "https://example.com"}
    ]


@mcp.tool(
    name="get_weather",
    description="Get Current weather for a city"
)
async def get_weather(city: Annotated[str, "City Name (e.g.Chicago)"]) -> Dict:
    return {
        "city": city,
        "temperature": 72,
        "condition": "Sunny",
        "humidity": 45,
    }


@mcp.tool()
def hello_world(name: str = "World") -> str:
    """Greet someone in a friendly way."""
    return f"Hello {name}, I'm an MCP server running in Python"


@mcp.tool()
def add(a: int, b: int) -> int:
    """Add 2 number together."""
    return a + b

# resources


@mcp.resource(uri="config://app")
def system_info() -> str:
    """Returns basic system information."""
    import platform
    return {
        "python_version": platform.python_version(),
        "os": platform.system(),
        "server_name": mcp.name,
    }


@mcp.resource(
    uri="file://notes/{note_id}"
)
async def get_note(note_id: str) -> str:
    """Retrieve a specific note by ID."""
    notes = {
        "meeting": "Q3 planning meeting notes...",
        "todo": "Finish MCP book chapter 4",
    }
    return notes.get(note_id, f"Note ID {note_id} not found")

# prompts

# registering a custom prompt


@mcp.prompt()
def research_template(topic: str) -> str:
    """High quality research workflow prompt."""
    return f"""
    You are a world-class researcher. Investigate the topic: {topic}
    """

# registering a static prompt


@mcp.prompt(
    name="Code Review",
    title="Code Review",
    description="Expert code review template"
)
def code_review(code: str) -> str:
    return (
        f"Review the following code for bugs, "
        f"style and performance issues...\n\n{code}"
    )


if __name__ == "__main__":
    # for local devs stdio
    # mcp.run()
    mcp.run(
        transport="streamable-http",
        # host="0.0.0.0",
        # port=8000,
    )
