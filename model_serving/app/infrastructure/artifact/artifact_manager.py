import hashlib
import os
from pathlib import Path

from domain.artifact.interface import ArtifactSource, ArtifactDownloader


class ArtifactManager(ArtifactSource):  # type: ignore
    def __init__(
        self,
        sources: dict[str, ArtifactDownloader],
        default_source: ArtifactDownloader,
        base_dir: str = "/tmp/artifacts",
    ):
        self._sources = sources
        self._default_source = default_source
        self._base_dir = Path(base_dir)

    def _resolve_path(self, url: str) -> str:
        h = hashlib.sha256(url.encode()).hexdigest()
        return str(self._base_dir / h)

    def _get_source(self, url: str) -> ArtifactDownloader:
        scheme = url.split("://")[0] if "://" in url else "file"
        return self._sources.get(scheme, self._default_source)

    async def download(self, url: str) -> str:  # type: ignore
        target_path = self._resolve_path(url)

        if os.path.exists(target_path):
            return target_path

        os.makedirs(os.path.dirname(target_path), exist_ok=True)

        source = self._get_source(url)

        return await source.download(url, target_path)
