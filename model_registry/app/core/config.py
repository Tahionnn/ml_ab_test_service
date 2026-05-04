from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    DATABASE_URL_MODEL_REGISTRY_SERVICE: str = ""
    KAFKA_BOOTSTRAP_SERVERS: str = Field(
        default="localhost:9092", alias="KAFKA_BOOTSTRAP_SERVERS"
    )

    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8")


settings = Settings()
