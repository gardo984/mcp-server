import os
import pathlib
import sys
from fastmcp import FastMCP

mcp = FastMCP("DesktopFileSystem")
current_directory = pathlib.Path(__file__).resolve().parent
BASE_DIR = (current_directory / 'desktop').resolve()
print(f"root dir: {BASE_DIR}")


def _validate_path(rel_path: str) -> pathlib.Path:
    target = (BASE_DIR / rel_path).resolve()
    if not str(target).startswith(str(BASE_DIR)):
        raise ValueError(f"Access denied: {rel_path}")
    return target


@mcp.tool()
def list_items() -> list[dict]:
    items = []
    for p in BASE_DIR.iterdir():
        items.append({
            "name": p.name,
            "is_dir": p.is_dir(),
        })
    return items


@mcp.tool()
def create(path: str, is_dir: bool = False) -> str:
    pass
