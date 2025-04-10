from fastapi import FastAPI, HTTPException
from loguru import logger

app = FastAPI(
    title="Hava Kalitesi Veri İşleme API",
    description="Bölgesel hava kalitesi verilerini sorgulama ve analiz API'si",
    version="1.0.0"
)

@app.get("/")
async def root():
    return {"message": "Hava Kalitesi Veri İşleme API'sine Hoş Geldiniz"}

@app.get("/health")
async def health_check():
    # Daha sonra servis bağlantı kontrolleri eklenecek
    return {"status": "healthy"}