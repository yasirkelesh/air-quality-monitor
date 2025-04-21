from fastapi import FastAPI, Request
from fastapi.responses import StreamingResponse
import asyncio
import json
import uvicorn  # Bu satırı ekleyin
from typing import AsyncGenerator, Set
from loguru import logger


from config import API_HOST, API_PORT

class SSEController:
    """Server-Sent Events controller"""
    
    def __init__(self):
        self.clients: Set[asyncio.Queue] = set()
        self.app = FastAPI(title="Anomali Bildirimleri SSE API")
        self.setup_routes()
    
    def setup_routes(self):
        """SSE endpoint'ini ayarla"""
        
        @self.app.get("/events")
        async def event_stream(request: Request):
            """SSE event stream endpoint'i"""
            queue = asyncio.Queue()
            self.clients.add(queue)
            
            async def event_generator() -> AsyncGenerator[str, None]:
                try:
                    while True:
                        if await request.is_disconnected():
                            break
                        
                        # Queue'dan mesaj al
                        data = await queue.get()
                        yield f"data: {json.dumps(data)}\n\n"
                        
                except asyncio.CancelledError:
                    logger.info("SSE bağlantısı iptal edildi")
                finally:
                    self.clients.remove(queue)
            
            return StreamingResponse(event_generator(), media_type="text/event-stream")
        
        @self.app.get("/health")
        async def health_check():
            return {"status": "healthy", "clients": len(self.clients)}
    
    async def broadcast_anomaly(self, anomaly: dict):
        """Tüm istemcilere anomali bildir"""
        # ObjectId'yi string'e çevir
        if "_id" in anomaly and not isinstance(anomaly["_id"], str):
            anomaly["_id"] = str(anomaly["_id"])
        
        # Tüm istemcilere gönder
        disconnected_clients = set()
        
        for queue in self.clients:
            try:
                await queue.put(anomaly)
            except Exception as e:
                logger.error(f"Anomali bildirimi gönderilemedi: {str(e)}")
                disconnected_clients.add(queue)
        
        # Bağlantısı kopan istemcileri temizle
        self.clients -= disconnected_clients
        
        if anomaly:
            logger.info(f"Anomali bildirimi gönderildi: {anomaly.get('source')} - {anomaly.get('pollutant')}")
    
    async def start_server(self):
        """SSE sunucusunu başlat"""
        config = uvicorn.Config(
            self.app,
            host=API_HOST,
            port=API_PORT,
            log_level="info"
        )
        server = uvicorn.Server(config)
        await server.serve()