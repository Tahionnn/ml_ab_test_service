import asyncio
import logging

from faststream.asgi import AsgiFastStream, make_ping_asgi

from bus.broker import broker
import bus.handlers  # noqa: F401


logger = logging.getLogger(__name__)

app = AsgiFastStream(
    broker,
    asyncapi_path="/asyncapi",
    asgi_routes=[
        ("/health", make_ping_asgi(broker, timeout=5.0)),
    ],
)


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    asyncio.run(app.run())
