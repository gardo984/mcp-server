import os
import pathlib
import sys
from fastmcp import FastMCP

mcp = FastMCP("DesktopFileSystem")
current_directory = pathlib.Path("/Users/mlazo")
BASE_DIR = (current_directory / 'Desktop').resolve()
print(f"root dir: {BASE_DIR}")


def _validate_path(rel_path: str) -> pathlib.Path:
    target = (BASE_DIR / rel_path).resolve()
    if not str(target).startswith(str(BASE_DIR)):
        raise ValueError(f"Access denied: {rel_path}")
    return target


@mcp.tool()
def list_items() -> list[dict]:
    """ List the available files on the local 'desktop' directory.

    Returns:
        items: A list of the available files.

        The returned list will have the following structure:
        {
            'name': 'file1.txt',
            'is_dir': False
        }
    """

    items = []
    for p in BASE_DIR.iterdir():
        items.append({
            "name": p.name,
            "is_dir": p.is_dir(),
        })
    return items


@mcp.tool()
def create(path: str, is_dir: bool = False) -> str:
    """ Create a file / directory on the local 'desktop' directory.

    Returns:
        str: Confirmation of the created file / directory

    Raises:
        ValueError: When the path can't be resolved, probably by permissions.
    """
    target = _validate_path(path)
    if is_dir:
        target.mkdir(parents=True, exist_ok=True)
        return f"Directory was created: {target}"
    else:
        target.parent.mkdir(parents=True, exist_ok=True)
        target.touch(exist_ok=True)
        return f"File was created: {target}"


@mcp.tool()
def append_to_file(path: str, content: str) -> str:
    """ Add a content into a file.

    Returns:
        str: Confirmation message of the added content into the path.

    Raises:
        FileNotFoundError: If the file does not exist.
        IsADirectoryError: If the specified path is a directory.
    """
    target = _validate_path(path)
    if not target.exists():
        raise FileNotFoundError(f"File {target} does not exist.")

    if target.is_dir():
        raise IsADirectoryError(f"Path {target} is a directory.")

    with open(target, 'a', encoding='utf-8') as f:
        f.write(content)
    return f"Content was added into the file {target}."


if __name__ == "__main__":
    # starts the mcp server
    mcp.run()
