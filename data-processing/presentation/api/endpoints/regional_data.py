from fastapi import APIRouter, Query, HTTPException
from typing import List, Dict, Any, Optional

from business.aggregation_service import AggregationService

router = APIRouter()

@router.get("/regional-averages", response_model=List[Dict[str, Any]])
async def get_regional_averages(
    hours: int = Query(24, description="Son kaç saatlik veri")
):
    """
    Tüm bölgeler için ortalama hava kalitesi değerlerini döndürür
    """
    try:
        aggregation_service = AggregationService()
        results = aggregation_service.get_all_regional_averages(hours)
        
        if not results:
            return []
        
        return results
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Bölgesel ortalamalar alınamadı: {str(e)}")

@router.get("/regional-average/{geohash}", response_model=Dict[str, Any])
async def get_regional_average(
    geohash: str,
    hours: int = Query(24, description="Son kaç saatlik veri")
):
    """
    Belirli bir geohash bölgesi için ortalama hava kalitesi değerlerini döndürür
    """
    try:
        aggregation_service = AggregationService()
        result = aggregation_service.get_regional_average(geohash, hours)
        
        if not result:
            raise HTTPException(status_code=404, detail=f"Geohash için veri bulunamadı: {geohash}")
        
        return result
    except HTTPException as e:
        raise e
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Bölgesel ortalama alınamadı: {str(e)}")