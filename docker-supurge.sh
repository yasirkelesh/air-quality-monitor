#!/bin/bash

echo "🧹 Docker süpürge çalışıyor..."

# Tüm container'ları durdur
echo "⏹️ Tüm container'lar durduruluyor..."
docker stop $(docker ps -aq) 2>/dev/null

# Tüm container'ları sil
echo "🗑️ Tüm container'lar siliniyor..."
docker rm $(docker ps -aq) 2>/dev/null

# Tüm image'leri sil
echo "🖼️ Tüm image'ler siliniyor..."
docker rmi -f $(docker images -aq) 2>/dev/null

# Tüm volume'ları sil
echo "📦 Tüm volume'lar siliniyor..."
docker volume rm $(docker volume ls -q) 2>/dev/null

# Tüm network'leri sil (default olanlar hariç)
echo "🌐 Custom Docker network'leri siliniyor..."
docker network rm $(docker network ls | grep -v "bridge\|host\|none" | awk '{ print $1 }') 2>/dev/null

# Sistem temizliği
echo "🧼 Docker sistem temizleniyor..."
docker system prune -af --volumes

echo "✅ Temizlik tamamlandı!"
