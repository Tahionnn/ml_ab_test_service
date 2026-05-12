from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    REDIS_URL: str = ""
    RAY_SERVE_ADDRESS: str = ""
    SERVE_URL: str = ""
    KAFKA_BOOTSTRAP_SERVERS: str = Field(
        default="localhost:9092", alias="KAFKA_BOOTSTRAP_SERVERS"
    )
    redis_url: str = Field(default="redis://localhost:6379/0", alias="REDIS_URL")
    artifacts_root: str = Field(
        default="/tmp/model-serving-artifacts", alias="ARTIFACTS_ROOT"
    )
    deployment_retry_attempts: int = Field(default=3, alias="DEPLOYMENT_RETRY_ATTEMPTS")
    deployment_retry_base_delay_sec: float = Field(
        default=3.0, alias="DEPLOYMENT_RETRY_BASE_DELAY_SEC"
    )
    deployment_ack_timeout_sec: int = Field(
        default=1800, alias="DEPLOYMENT_ACK_TIMEOUT_SEC"
    )
    default_consumer_group: str = Field(
        default="model-serving-workers", alias="KAFKA_CONSUMER_GROUP"
    )
    default_consumer_name: str = Field(default="worker-1", alias="KAFKA_CONSUMER_NAME")

    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8")


settings = Settings()
