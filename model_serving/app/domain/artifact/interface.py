from typing import Protocol


class ArtifactSource(Protocol):
    async def download(self, artifatc_url: str) -> str: ...


class ArtifactDownloader(Protocol):
    async def download(self, url: str, target_path: str) -> str: ...
