// GSM (Global Scalable Matrix) Real-time Control Center Logic

const API_BASE = window.location.origin; // Aynı sunucu üzerinden servis ediliyor
let updateInterval = null;

// DOM Elemanları
const systemStatusBadge = document.getElementById('system-status');
const metricCpu = document.getElementById('metric-cpu');
const metricMem = document.getElementById('metric-mem');
const metricNodes = document.getElementById('metric-nodes');
const metricPing = document.getElementById('metric-ping');
const cpuBar = document.getElementById('cpu-bar');
const memBar = document.getElementById('mem-bar');
const streamContainer = document.getElementById('event-stream-container');
const eventTypeInput = document.getElementById('event-type');
const eventPayloadInput = document.getElementById('event-payload');
const btnPublish = document.getElementById('btn-publish');
const btnClearFeed = document.getElementById('btn-clear-feed');

// Başlangıç Kurulumları
document.addEventListener('DOMContentLoaded', () => {
    checkSystemStatus();
    fetchEvents();
    
    // Düzenli olarak metrikleri ve olayları güncelle (Asenkron Döngü)
    updateInterval = setInterval(() => {
        checkSystemStatus();
        fetchEvents();
    }, 2500);

    // Event Dinleyicileri
    btnPublish.addEventListener('click', injectEvent);
    btnClearFeed.addEventListener('click', () => {
        streamContainer.innerHTML = '<div class="stream-placeholder">Stream cleared. Waiting for events...</div>';
    });
});

// Sistem Bağlantı Durumu Kontrolü
async function checkSystemStatus() {
    const startTime = Date.now();
    try {
        const response = await fetch(`${API_BASE}/api/status`);
        const duration = Date.now() - startTime;
        
        if (response.ok) {
            const data = await response.json();
            systemStatusBadge.textContent = "OPERATIONAL";
            systemStatusBadge.className = "status-badge connected";
            metricPing.textContent = `${duration} ms`;
        } else {
            setDisconnectedState();
        }
    } catch (error) {
        setDisconnectedState();
    }
}

function setDisconnectedState() {
    systemStatusBadge.textContent = "OFFLINE";
    systemStatusBadge.className = "status-badge";
    metricPing.textContent = "-- ms";
}

// Olayları Çekme ve Ekrana Basma
async function fetchEvents() {
    try {
        const response = await fetch(`${API_BASE}/api/events`);
        if (!response.ok) return;

        const events = await response.json();
        if (events && events.length > 0) {
            renderEventStream(events);
            // En son sistem metriğini yakalayıp kartları güncelle
            const systemTicks = events.filter(e => e.type === 'SYSTEM_METRIC_TICK');
            if (systemTicks.length > 0) {
                const latestTick = systemTicks[systemTicks.length - 1];
                updateMetricsUI(latestTick.payload);
            }
        }
    } catch (error) {
        console.error("Error fetching matrix events:", error);
    }
}

// Metrik Kartlarının Güncellenmesi (Mikro-Animasyon ve Değerler)
function updateMetricsUI(payload) {
    if (!payload) return;
    
    const cpuVal = payload.cpu_usage || 0;
    const memVal = payload.memory_free_gb || 0;
    const nodesVal = payload.active_nodes || 0;

    metricCpu.textContent = `${cpuVal.toFixed(1)}%`;
    cpuBar.style.width = `${cpuVal}%`;

    // 16GB üzerinden doluluk hesaplayalım (örnek)
    const memPercent = ((16 - memVal) / 16) * 100;
    metricMem.textContent = `${(16 - memVal).toFixed(1)} / 16 GB`;
    memBar.style.width = `${memPercent}%`;

    metricNodes.textContent = nodesVal;
}

// Canlı Akış Renderlama
function renderEventStream(events) {
    // Mevcut listenin son halini render edelim
    streamContainer.innerHTML = '';
    
    events.slice().reverse().forEach(event => {
        const logCard = document.createElement('div');
        let logClass = 'event-log';
        if (event.type.includes('SYSTEM')) {
            logClass += ' system';
        } else {
            logClass += ' custom';
        }
        
        logCard.className = logClass;
        
        const timestamp = new Date(event.timestamp).toLocaleTimeString();
        
        logCard.innerHTML = `
            <div class="log-meta">
                <span class="log-type">${event.type}</span>
                <span class="log-time">${timestamp}</span>
            </div>
            <div class="log-payload">${JSON.stringify(event.payload, null, 2)}</div>
        `;
        streamContainer.appendChild(logCard);
    });
}

// Yeni Olay Enjekte Etme
async function injectEvent() {
    const type = eventTypeInput.value.trim();
    let payload = {};
    
    try {
        payload = JSON.parse(eventPayloadInput.value);
    } catch (e) {
        alert("Payload must be valid JSON!");
        return;
    }

    if (!type) {
        alert("Event type is required!");
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/api/publish`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ type, payload })
        });

        if (response.ok) {
            // Butona basıldığında mikro-animasyon tetikle
            btnPublish.style.transform = 'scale(0.95)';
            setTimeout(() => btnPublish.style.transform = 'none', 100);
            
            // Olayları anında tazele
            fetchEvents();
        } else {
            alert("Failed to inject event into Matrix");
        }
    } catch (error) {
        console.error("Error injecting event:", error);
    }
}
