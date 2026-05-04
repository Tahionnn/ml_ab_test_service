from faststream.kafka.fastapi import KafkaRouter

from app.core.config import settings

kafka_router = KafkaRouter(
    settings.KAFKA_BOOTSTRAP_SERVERS,
    schema_url="/asyncapi",
    include_in_schema=True,
)
