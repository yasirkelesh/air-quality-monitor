from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from loguru import logger

from presentation.api.endpoints import regional_data

app = FastAPI(
    title="Hava Kalitesi Veri İşleme API",
    description="Bölgesel hava kalitesi verilerini sorgulama ve analiz API'si",
    version="1.0.0"
)

# CORS ayarları
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Endpoint'leri ekle
app.include_router(regional_data.router, tags=["Bölgesel Veriler"])

@app.get("/")
async def root():
    return {"message": "Hava Kalitesi Veri İşleme API'sine Hoş Geldiniz"}

@app.get("/health")
async def health_check():
    # TODO: Servis bağlantı kontrolleri eklenecek
    return {"status": "healthy"}