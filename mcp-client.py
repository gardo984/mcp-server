
import sys
import asyncio
from dotenv import load_dotenv
from anthropic import Anthropic
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client

load_dotenv()

# define model
MODEL = 'claude-3-5-sonnet-20241022'
MAX_TOKENS = 1000


async def list_tools(session):
    """ Get the list of available tools """

    outcome = await session.list_tools()
    return [
        dict(
            name=item.name,
            description=item.description,
            input_schema=item.inputSchema,
        )
        for item in outcome.tools
    ]


def llm_call(
    anth, messages, tools=[], system=None,
):
    params = dict(
        model=MODEL,
        messages=messages,
        max_tokens=MAX_TOKENS,
    )
    if tools:
        params.update({
            "tools": tools,
        })
    if system:
        params.update({
            "system": system,
        })
    return anth.messages.create(**params)


async def handle_tool(
    part, session, anth, messages, llm_text,
):
    prompt = getattr(part, "text", "")
    if prompt:
        llm_text.append(prompt)
        messages.append(
            dict(role="assistant", content=prompt)
        )
    result = await session.call_tool(
        part.name, part.input,
    )

    if isinstance(result.content, str):
        user_input = result.content
    else:
        user_input = ''.join([
            getattr(item, "text", str(item))
            for item in result.content
        ])
    messages.append(dict(role="user", content=user_input))
    followup = llm_call(anth, messages)
    for output in followup.content:
        if output.type == "text":
            llm_text.append(output.text)
            messages.append(
                dict(role="assistant", content=output.text)
            )


async def main(
    server_script: str,
    query: str,
):
    cmd = "python" if server_script.endswith(".py") else "node"
    params = StdioServerParameters(
        command=cmd, args=[server_script]
    )

    async with stdio_client(params) as (stdio, write),
    ClientSession(stdio, write) as session:
        await session.initialize()

        tools = await list_tools(session)
        anth = Anthropic()
        messages = [dict(role="user", content=query)]
        ai_resp = llm_call(anth, messages, tools=tools)
        llm_text = []
        for part in ai_resp.content:
            if part.type == "text":
                llm_text.append(part.text)
                messages.append(
                    dict(role="assistant", content=part.text)
                )
            elif part.type == "tool_use":
                await handle_tool(
                    part, session, anth, messages, llm_text
                )
            print(llm_text[-1])

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python client.py <server-path.py|js> <query>")
        sys.exit(1)

    asyncio.run(main(
        sys.argv[1], sys.argv[2]
    ))

