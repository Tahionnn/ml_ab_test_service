from domain.artifact.interface import ArtifactDownloader


class MockArtifactSource(ArtifactDownloader):  # type: ignore

    async def download(self, url: str, target_path: str) -> str:
        with open(target_path, "w") as f:
            f.write(f"weights_for_model_from_{url}")
        return target_path
