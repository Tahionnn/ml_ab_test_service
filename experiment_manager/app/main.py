from contextlib import asynccontextmanager
from collections.abc import AsyncGenerator
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from core.database import create_tables
from core.exceptions import register_exceptions

from api.v1.experiments import experiment_router
from api.v1.variants import variant_router
from api.v1.metrics import metric_router
from api.v1.internal import internal_router


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    await create_tables()
    yield


app = FastAPI(
    lifespan=lifespan,
    redoc_url=None,
)


routers = (experiment_router, variant_router, metric_router, internal_router)

for router in routers:
    app.include_router(router)

register_exceptions(app)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)
