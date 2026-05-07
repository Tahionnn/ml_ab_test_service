from faststream.kafka import KafkaBroker

from app.core.config import settings


broker = KafkaBroker(
    settings.KAFKA_BOOTSTRAP_SERVERS,
    # enable_idempotence=True
)
